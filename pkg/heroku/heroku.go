package heroku

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HerokuProvider represents a client for interacting with the Heroku API
type HerokuProvider struct {
	APIKey string
	Client *http.Client
}

// NewHerokuProvider creates a new HerokuProvider with the given API key
func NewHerokuProvider(apiKey string) *HerokuProvider {
	return &HerokuProvider{
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

// GetAllAppsConfig retrieves the configuration for all Heroku apps
func (h *HerokuProvider) GetAllAppsConfig() ([]map[string]interface{}, error) {
	url := "https://api.heroku.com/apps"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", h.APIKey))
	req.Header.Add("Accept", "application/vnd.heroku+json; version=3")

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	var apps []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&apps)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return apps, nil
}
