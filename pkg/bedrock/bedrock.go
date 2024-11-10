package bedrock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/smithy-go"
)

// RateLimiter handles API request rate limiting
type RateLimiter struct {
	tokens   chan struct{}
	interval time.Duration
	mu       sync.Mutex
}

// newRateLimiter creates a new rate limiter with specified requests per interval
func newRateLimiter(maxRequests int, interval time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		tokens:   make(chan struct{}, maxRequests),
		interval: interval,
	}

	// Initialize token bucket
	for i := 0; i < maxRequests; i++ {
		limiter.tokens <- struct{}{}
	}

	// Replenish tokens periodically
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			limiter.mu.Lock()
			currentTokens := len(limiter.tokens)
			needed := maxRequests - currentTokens
			for i := 0; i < needed; i++ {
				select {
				case limiter.tokens <- struct{}{}:
				default:
					// Channel is full
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

func (r *RateLimiter) wait() {
	<-r.tokens
}

func (r *RateLimiter) release() {
	r.mu.Lock()
	defer r.mu.Unlock()
	select {
	case r.tokens <- struct{}{}:
	default:
		// Channel is full
	}
}

// BedrockClient represents a client for interacting with AWS Bedrock
type BedrockClient struct {
	bedrockClient *bedrockruntime.Client
	rateLimiter   *RateLimiter
	config        ClientConfig
	semaphore     chan struct{} // Added for parallel request limiting
}

// ClientConfig holds configuration options for the Bedrock client
type ClientConfig struct {
	MaxRequestsPerMinute int
	MaxRetries           int
	InitialRetryDelay    time.Duration
	MaxRetryDelay        time.Duration
	AWSRegion            string
	InferenceProfileARN  string
	MaxParallelRequests  int
}

// DefaultConfig returns the default client configuration
func DefaultConfig() ClientConfig {
	return ClientConfig{
		MaxRequestsPerMinute: 50,              // 50 requests per minute by default
		MaxRetries:           20,              // Maximum number of retries
		InitialRetryDelay:    1 * time.Second, // Start with 1 second delay
		MaxRetryDelay:        3 * time.Minute, // Maximum delay between retries
		AWSRegion:            "us-east-1",     // Default AWS region
		InferenceProfileARN:  "",              // Must be set by user
		MaxParallelRequests:  5,               // Default to 5 parallel requests
	}
}

// BedrockRequest represents the request structure for Bedrock API
type BedrockRequest struct {
	AnthropicVersion string    `json:"anthropic_version"`
	Messages         []Message `json:"messages"`
	MaxTokens        int       `json:"max_tokens"`
	Temperature      float64   `json:"temperature"`
	TopP             float64   `json:"top_p"`
	TopK             int       `json:"top_k"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// BedrockResponse represents the response structure from Bedrock API
type BedrockResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// NewBedrockClient creates a new BedrockClient with AWS credentials and optional config
func NewBedrockClient(awsKey string, awsSecret string, config ...ClientConfig) (*BedrockClient, error) {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.InferenceProfileARN == "" {
		return nil, fmt.Errorf("InferenceProfileARN is required")
	}

	if cfg.MaxParallelRequests < 1 {
		return nil, fmt.Errorf("MaxParallelRequests must be at least 1")
	}

	// Load AWS configuration with credentials
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsKey, awsSecret, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	// Create Bedrock client
	client := bedrockruntime.NewFromConfig(awsCfg)

	return &BedrockClient{
		bedrockClient: client,
		rateLimiter:   newRateLimiter(cfg.MaxRequestsPerMinute, time.Minute),
		config:        cfg,
		semaphore:     make(chan struct{}, cfg.MaxParallelRequests),
	}, nil
}

// Messages sends a chat request to Claude AI via AWS Bedrock and returns the response
func (c *BedrockClient) Messages(prompt string) (string, error) {
	// Acquire semaphore slot
	c.semaphore <- struct{}{}
	defer func() {
		<-c.semaphore // Release semaphore slot
	}()

	request := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   8192,
		Temperature: 0.7,
		TopP:        1,
		TopK:        250,
	}

	jsonPayload, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling payload: %w", err)
	}

	var content string
	retryDelay := c.config.InitialRetryDelay

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		c.rateLimiter.wait()

		input := &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(c.config.InferenceProfileARN),
			Accept:      aws.String("application/json"),
			ContentType: aws.String("application/json"),
			Body:        jsonPayload,
		}

		output, err := c.bedrockClient.InvokeModel(context.Background(), input)

		if err != nil {
			c.rateLimiter.release()

			var apiErr smithy.APIError
			if ok := errors.As(err, &apiErr); ok {
				switch apiErr.ErrorCode() {
				case "ThrottlingException", "ServiceUnavailable", "InternalServerError":
					if attempt < c.config.MaxRetries-1 {
						jitter := time.Duration(rand.Int63n(int64(retryDelay) / 2))
						sleepTime := retryDelay + jitter

						if sleepTime > c.config.MaxRetryDelay {
							sleepTime = c.config.MaxRetryDelay
						}

						log.Printf("Request failed with error %s. Retrying in %v (attempt %d/%d)",
							apiErr.ErrorCode(), sleepTime, attempt+1, c.config.MaxRetries)

						time.Sleep(sleepTime)
						retryDelay *= 2 // Exponential backoff
						continue
					}
				}
			}
			return "", fmt.Errorf("error invoking model: %w", err)
		}

		var response BedrockResponse
		if err := json.Unmarshal(output.Body, &response); err != nil {
			c.rateLimiter.release()
			return "", fmt.Errorf("error decoding response: %w", err)
		}

		c.rateLimiter.release()

		if len(response.Content) > 0 {
			content = response.Content[0].Text
			return content, nil
		}

		return "", fmt.Errorf("empty response from model")
	}

	return "", fmt.Errorf("max retries reached without successful response")
}
