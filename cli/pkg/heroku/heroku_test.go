package heroku

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHerokuProvider_GetAllAppsConfig(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.heroku+json; version=3", r.Header.Get("Accept"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
            {
                "name": "app1",
                "region": {"name": "us"},
                "stack": {"name": "heroku-18"}
            },
            {
                "name": "app2",
                "region": {"name": "eu"},
                "stack": {"name": "heroku-20"}
            }
        ]`))
	}))
	defer server.Close()

	// Create a HerokuProvider with the mock server URL
	provider := &HerokuProvider{
		APIKey: "test-api-key",
		Client: server.Client(),
	}

	// Test the GetAllAppsConfig method
	apps, err := provider.GetAllAppsConfig()
	assert.NoError(t, err)
	assert.Len(t, apps, 2)
	assert.Equal(t, "app1", apps[0]["name"])
	assert.Equal(t, "app2", apps[1]["name"])
}
