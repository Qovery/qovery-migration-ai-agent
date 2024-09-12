package claude

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaudeClient_Chat(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
            "choices": [
                {
                    "message": {
                        "content": "Hello, this is Claude!"
                    }
                }
            ]
        }`))
	}))
	defer server.Close()

	// Create a ClaudeClient with the mock server URL
	client := &ClaudeClient{
		APIKey: "test-api-key",
		Client: server.Client(),
	}

	// Test the Messages method
	response, err := client.Messages("Hello, Claude!")
	assert.NoError(t, err)
	assert.Equal(t, "Hello, this is Claude!", response)
}
