package sources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	cleverCloudAPIRootURL = "https://api.clever-cloud.com/v2"
	authURL               = "https://api.clever-cloud.com/v2/oauth/token"
)

// CleverCloudProvider represents a client for interacting with the Clever Cloud API
type CleverCloudProvider struct {
	ConsumerKey    string
	ConsumerSecret string
	Token          string
	Secret         string
	AccessToken    string
	Client         *http.Client
}

// CleverCloudAppConfig represents the configuration for a Clever Cloud app, including costs, addons, and domains
type CleverCloudAppConfig struct {
	mApp      map[string]interface{}
	Config    map[string]string
	Addons    []map[string]interface{}
	Domains   []map[string]interface{}
	TotalCost float64
}

func (a CleverCloudAppConfig) App() map[string]interface{} {
	return a.mApp
}

func (a CleverCloudAppConfig) Cost() float64 {
	return a.TotalCost
}

// Map returns a map representation of the AppConfig
func (a CleverCloudAppConfig) Map() map[string]interface{} {
	return map[string]interface{}{
		"app":     a.App,
		"config":  a.Config,
		"addons":  a.Addons,
		"domains": a.Domains,
		"cost":    a.Cost,
	}
}

func (a CleverCloudAppConfig) Name() string {
	appName, _ := a.mApp["name"].(string)
	return appName
}

// CleverCloudError represents an error returned by the Clever Cloud API
type CleverCloudError struct {
	Message string `json:"message"`
}

// NewCleverCloudProvider creates a new CleverCloudProvider with the given credentials
func NewCleverCloudProvider(consumerKey, consumerSecret, token, secret string) (*CleverCloudProvider, error) {
	provider := &CleverCloudProvider{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		Token:          token,
		Secret:         secret,
		Client:         &http.Client{},
	}

	err := provider.authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return provider, nil
}

func (c *CleverCloudProvider) authenticate() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("error creating auth request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.ConsumerKey, c.ConsumerSecret)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending auth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading auth response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
	}
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		return fmt.Errorf("error decoding auth response: %w", err)
	}

	c.AccessToken = authResp.AccessToken
	return nil
}

// GetAllAppsConfig retrieves the configuration for all Clever Cloud apps, including env vars, addons, domains, and costs
func (c *CleverCloudProvider) GetAllAppsConfig() ([]AppConfig, error) {
	apps, err := c.getApps()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	configs := make([]AppConfig, len(apps))

	for i, app := range apps {
		wg.Add(1)
		go func(i int, app map[string]interface{}) {
			defer wg.Done()
			appID, _ := app["id"].(string)
			config, err := c.getAppConfig(appID)
			if err != nil {
				fmt.Printf("Error fetching config for app %s: %v\n", appID, err)
				return
			}
			addons, err := c.getAppAddons(appID)
			if err != nil {
				fmt.Printf("Error fetching addons for app %s: %v\n", appID, err)
				return
			}
			domains, err := c.getAppDomains(appID)
			if err != nil {
				fmt.Printf("Error fetching domains for app %s: %v\n", appID, err)
				return
			}
			cost, err := c.getAppCost(appID)
			if err != nil {
				fmt.Printf("Error fetching cost for app %s: %v\n", appID, err)
				return
			}

			configs[i] = CleverCloudAppConfig{
				mApp:      app,
				Config:    config,
				Addons:    addons,
				Domains:   domains,
				TotalCost: cost,
			}
		}(i, app)
	}

	wg.Wait()

	return configs, nil
}

func (c *CleverCloudProvider) getApps() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/organisations/orga_/applications", cleverCloudAPIRootURL)
	return c.makeRequest("GET", url, nil)
}

func (c *CleverCloudProvider) getAppConfig(appID string) (map[string]string, error) {
	url := fmt.Sprintf("%s/applications/%s/environment", cleverCloudAPIRootURL, appID)
	return c.makeRequestConfig("GET", url)
}

func (c *CleverCloudProvider) getAppAddons(appID string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/applications/%s/addons", cleverCloudAPIRootURL, appID)
	return c.makeRequest("GET", url, nil)
}

func (c *CleverCloudProvider) getAppDomains(appID string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/applications/%s/vhosts", cleverCloudAPIRootURL, appID)
	return c.makeRequest("GET", url, nil)
}

func (c *CleverCloudProvider) getAppCost(appID string) (float64, error) {
	// Note: Clever Cloud API doesn't seem to have a direct endpoint for app cost
	// This is a placeholder implementation. You may need to calculate the cost
	// based on the app's configuration and pricing information.
	return 0, nil
}

func (c *CleverCloudProvider) makeRequest(method, url string, body []byte) ([]map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, ioutil.NopCloser(strings.NewReader(string(body))))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.addAuthHeader(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var cleverErr CleverCloudError
		if err := json.Unmarshal(respBody, &cleverErr); err == nil {
			return nil, fmt.Errorf("API error: %s", cleverErr.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result []map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		// If unmarshaling to slice fails, try unmarshaling to single object
		var singleResult map[string]interface{}
		if err := json.Unmarshal(respBody, &singleResult); err == nil {
			result = []map[string]interface{}{singleResult}
		} else {
			return nil, fmt.Errorf("error decoding response: %w", err)
		}
	}

	return result, nil
}

func (c *CleverCloudProvider) makeRequestConfig(method, url string) (map[string]string, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.addAuthHeader(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var cleverErr CleverCloudError
		if err := json.Unmarshal(body, &cleverErr); err == nil {
			return nil, fmt.Errorf("API error: %s", cleverErr.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]string
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return result, nil
}

func (c *CleverCloudProvider) addAuthHeader(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Add("Accept", "application/json")
}
