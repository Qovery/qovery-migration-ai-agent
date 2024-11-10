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
}

// ClientConfig holds configuration options for the Bedrock client
type ClientConfig struct {
	MaxRequestsPerMinute int
	MaxRetries           int
	InitialRetryDelay    time.Duration
	MaxRetryDelay        time.Duration
	AWSRegion            string
}

// DefaultConfig returns the default client configuration
func DefaultConfig() ClientConfig {
	return ClientConfig{
		MaxRequestsPerMinute: 50,              // 50 requests per minute by default
		MaxRetries:           20,              // Maximum number of retries
		InitialRetryDelay:    1 * time.Second, // Start with 1 second delay
		MaxRetryDelay:        3 * time.Minute, // Maximum delay between retries
		AWSRegion:            "us-east-1",     // Default AWS region
	}
}

// BedrockRequest represents the request structure for Bedrock API
type BedrockRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature"`
	TopP              float64  `json:"top_p"`
	TopK              int      `json:"top_k"`
	StopSequences     []string `json:"stop_sequences"`
}

// BedrockResponse represents the response structure from Bedrock API
type BedrockResponse struct {
	Completion string `json:"completion"`
}

// NewBedrockClient creates a new BedrockClient with AWS credentials and optional config
func NewBedrockClient(awsKey string, awsSecret string, config ...ClientConfig) *BedrockClient {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Load AWS configuration with credentials
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsKey, awsSecret, "")),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}

	// Create Bedrock client
	client := bedrockruntime.NewFromConfig(awsCfg)

	return &BedrockClient{
		bedrockClient: client,
		rateLimiter:   newRateLimiter(cfg.MaxRequestsPerMinute, time.Minute),
		config:        cfg,
	}
}

// Messages sends a chat request to Claude AI via AWS Bedrock and returns the response
func (c *BedrockClient) Messages(prompt string) (string, error) {
	request := BedrockRequest{
		Prompt:            prompt,
		MaxTokensToSample: 8192,
		Temperature:       0.7,
		TopP:              1,
		TopK:              250,
		StopSequences:     []string{"\n\nHuman:", "\n\nAssistant:"},
	}

	jsonPayload, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling payload: %w", err)
	}

	var content string
	retryDelay := c.config.InitialRetryDelay

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		// Wait for rate limit token
		c.rateLimiter.wait()

		// Create Bedrock API request
		input := &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String("anthropic.claude-3-5-sonnet-20241022-v2:0"),
			Accept:      aws.String("application/json"),
			ContentType: aws.String("application/json"),
			Body:        jsonPayload,
		}

		// Invoke the model
		output, err := c.bedrockClient.InvokeModel(context.Background(), input)

		if err != nil {
			c.rateLimiter.release()

			// Handle specific AWS errors
			var apiErr smithy.APIError
			if ok := errors.As(err, &apiErr); ok {
				switch apiErr.ErrorCode() {
				case "ThrottlingException", "ServiceUnavailable", "InternalServerError":
					if attempt < c.config.MaxRetries-1 {
						// Add jitter to prevent thundering herd
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
		content = response.Completion
		return content, nil
	}

	return "", fmt.Errorf("max retries reached without successful response")
}
