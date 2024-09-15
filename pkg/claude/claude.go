package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// ClaudeClient represents a client for interacting with the Claude AI API
type ClaudeClient struct {
	APIKey string
	Client *http.Client
}

// NewClaudeClient creates a new ClaudeClient with the given API key
func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

// Messages sends a chat request to the Claude AI API and returns the response
func (c *ClaudeClient) Messages(prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"
	payload := map[string]interface{}{
		"model": "claude-3-5-sonnet-20240620",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 8192,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	var content string

	// check that the Claude API response contains the expected content
	if _, ok := result["content"]; !ok {
		return "", fmt.Errorf("error: response does not contain content. Here is the full response: %v", result)
	}

	for _, m := range result["content"].([]interface{}) {
		x := m.(map[string]interface{})
		content += x["text"].(string)
	}

	return content, nil
}
