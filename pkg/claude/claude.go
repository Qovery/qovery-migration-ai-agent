package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
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

// wait blocks until a token is available
func (r *RateLimiter) wait() {
	<-r.tokens
}

// release returns a token to the bucket
func (r *RateLimiter) release() {
	r.mu.Lock()
	defer r.mu.Unlock()
	select {
	case r.tokens <- struct{}{}:
	default:
		// Channel is full
	}
}

// ClaudeClient represents a client for interacting with the Claude AI API
type ClaudeClient struct {
	APIKey      string
	Client      *http.Client
	rateLimiter *RateLimiter
	config      ClientConfig
}

// ClientConfig holds configuration options for the Claude client
type ClientConfig struct {
	MaxRequestsPerMinute int
	MaxRetries           int
	InitialRetryDelay    time.Duration
	MaxRetryDelay        time.Duration
}

// DefaultConfig returns the default client configuration
func DefaultConfig() ClientConfig {
	return ClientConfig{
		MaxRequestsPerMinute: 50,               // 50 requests per minute by default
		MaxRetries:           10,               // Maximum number of retries
		InitialRetryDelay:    1 * time.Second,  // Start with 1 second delay
		MaxRetryDelay:        60 * time.Second, // Maximum delay between retries
	}
}

// NewClaudeClient creates a new ClaudeClient with the given API key and optional config
func NewClaudeClient(apiKey string, config ...ClientConfig) *ClaudeClient {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &ClaudeClient{
		APIKey:      apiKey,
		Client:      &http.Client{},
		rateLimiter: newRateLimiter(cfg.MaxRequestsPerMinute, time.Minute),
		config:      cfg,
	}
}

// Messages sends a chat request to the Claude AI API and returns the response
func (c *ClaudeClient) Messages(prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"
	payload := map[string]interface{}{
		"model": "claude-3-5-sonnet-20241022",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 8192,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling payload: %w", err)
	}

	var content string
	retryDelay := c.config.InitialRetryDelay

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		// Wait for rate limit token
		c.rateLimiter.wait()

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			c.rateLimiter.release()
			return "", fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", c.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		resp, err := c.Client.Do(req)
		if err != nil {
			c.rateLimiter.release()
			return "", fmt.Errorf("error sending request: %w", err)
		}

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if err != nil {
			c.rateLimiter.release()
			return "", fmt.Errorf("error decoding response: %w", err)
		}

		// Handle different types of errors
		switch resp.StatusCode {
		case http.StatusOK:
			c.rateLimiter.release()
			// Process successful response
			if _, ok := result["content"]; !ok {
				return "", fmt.Errorf("error: response does not contain content. Full response: %v", result)
			}

			for _, m := range result["content"].([]interface{}) {
				x := m.(map[string]interface{})
				content += x["text"].(string)
			}
			return content, nil

		case http.StatusTooManyRequests, http.StatusServiceUnavailable, http.StatusInternalServerError:
			// Release token for rate-limited requests
			c.rateLimiter.release()

			if attempt < c.config.MaxRetries-1 {
				// Add jitter to prevent thundering herd
				jitter := time.Duration(rand.Int63n(int64(retryDelay) / 2))
				sleepTime := retryDelay + jitter

				// Ensure we don't exceed max retry delay
				if sleepTime > c.config.MaxRetryDelay {
					sleepTime = c.config.MaxRetryDelay
				}

				log.Printf("Request failed with status code %d. Retrying in %v (attempt %d/%d)",
					resp.StatusCode, sleepTime, attempt+1, c.config.MaxRetries)

				time.Sleep(sleepTime)
				retryDelay *= 2 // Exponential backoff
				continue
			}
			return "", fmt.Errorf("max retries reached. Last error: %v", result)

		default:
			// Release token for non-retryable errors
			c.rateLimiter.release()
			return "", fmt.Errorf("request failed with status code %d: %v", resp.StatusCode, result)
		}
	}

	return "", fmt.Errorf("max retries reached without successful response")
}
