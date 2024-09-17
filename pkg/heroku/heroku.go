package heroku

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	herokuAPIRootURL = "https://api.heroku.com"
)

// HerokuProvider represents a client for interacting with the Heroku API
type HerokuProvider struct {
	APIKey string
	Client *http.Client
}

// AppConfig represents the configuration for a Heroku app, including costs, pipeline info, and review apps
type AppConfig struct {
	App           map[string]interface{}
	Config        map[string]string
	Addons        []map[string]interface{}
	Domains       []map[string]interface{}
	Cost          float64
	Pipeline      map[string]interface{}
	ReviewApps    []map[string]interface{}
	ReviewAppConf map[string]interface{}
}

// Map returns a map representation of the AppConfig
func (a AppConfig) Map() map[string]interface{} {
	return map[string]interface{}{
		"app":             a.App,
		"config":          a.Config,
		"addons":          a.Addons,
		"domains":         a.Domains,
		"cost":            a.Cost,
		"pipeline":        a.Pipeline,
		"review_apps":     a.ReviewApps,
		"review_app_conf": a.ReviewAppConf,
	}
}

// HerokuError represents an error returned by the Heroku API
type HerokuError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// NewHerokuProvider creates a new HerokuProvider with the given API key
func NewHerokuProvider(apiKey string) *HerokuProvider {
	return &HerokuProvider{
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

// GetAllAppsConfig retrieves the configuration for all Heroku apps, including env vars, addons, domains, costs, pipeline info, and review apps
func (h *HerokuProvider) GetAllAppsConfig() ([]AppConfig, error) {
	apps, err := h.getApps()
	if err != nil {
		return nil, err
	}

	pipelines, err := h.getPipelines()
	if err != nil {
		return nil, err
	}

	pipelineMap := make(map[string]map[string]interface{})
	for _, pipeline := range pipelines {
		pipelineID, _ := pipeline["id"].(string)
		pipelineMap[pipelineID] = pipeline
	}

	var wg sync.WaitGroup
	configs := make([]AppConfig, len(apps))

	for i, app := range apps {
		wg.Add(1)
		go func(i int, app map[string]interface{}) {
			defer wg.Done()
			appName, _ := app["name"].(string)
			config, err := h.getAppConfig(appName)
			if err != nil {
				fmt.Printf("Error fetching config for app %s: %v\n", appName, err)
				return
			}
			addons, err := h.getAppAddons(appName)
			if err != nil {
				fmt.Printf("Error fetching addons for app %s: %v\n", appName, err)
				return
			}
			domains, err := h.getAppDomains(appName)
			if err != nil {
				fmt.Printf("Error fetching domains for app %s: %v\n", appName, err)
				return
			}
			cost, err := h.getAppCost(appName)
			if err != nil {
				fmt.Printf("Error fetching cost for app %s: %v\n", appName, err)
				return
			}
			pipelineCoupling, err := h.getAppPipelineCoupling(appName)
			if err != nil {
				fmt.Printf("Error fetching pipeline coupling for app %s: %v\n", appName, err)
				return
			}

			var pipeline map[string]interface{}
			var reviewApps []map[string]interface{}
			var reviewAppConf map[string]interface{}
			if pipelineCoupling != nil {
				pipelineID, ok := pipelineCoupling["pipeline"].(map[string]interface{})["id"].(string)
				if ok {
					pipeline = pipelineMap[pipelineID]
					reviewApps, err = h.getPipelineReviewApps(pipelineID)
					if err != nil {
						fmt.Printf("Error fetching review apps for pipeline %s: %v\n", pipelineID, err)
					}
					reviewAppConf, err = h.getPipelineReviewAppConfig(pipelineID)
					if err != nil {
						fmt.Printf("Error fetching review app config for pipeline %s: %v\n", pipelineID, err)
					}
				}
			}

			configs[i] = AppConfig{
				App:           app,
				Config:        config,
				Addons:        addons,
				Domains:       domains,
				Cost:          cost,
				Pipeline:      pipeline,
				ReviewApps:    reviewApps,
				ReviewAppConf: reviewAppConf,
			}
		}(i, app)
	}

	wg.Wait()

	return configs, nil
}

func (h *HerokuProvider) getApps() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/apps", herokuAPIRootURL)
	return h.makeRequest(url)
}

func (h *HerokuProvider) getAppConfig(appName string) (map[string]string, error) {
	url := fmt.Sprintf("%s/apps/%s/config-vars", herokuAPIRootURL, appName)
	return h.makeRequestConfig(url)
}

func (h *HerokuProvider) getAppAddons(appName string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/apps/%s/addons", herokuAPIRootURL, appName)
	return h.makeRequest(url)
}

func (h *HerokuProvider) getAppDomains(appName string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/apps/%s/domains", herokuAPIRootURL, appName)
	return h.makeRequest(url)
}

func (h *HerokuProvider) getAppCost(appName string) (float64, error) {
	now := time.Now()
	startOfPeriod := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	url := fmt.Sprintf("%s/apps/%s/formation", herokuAPIRootURL, appName)

	formations, err := h.makeRequest(url)
	if err != nil {
		return 0, err
	}

	var totalCost float64
	daysInPeriod := float64(now.Sub(startOfPeriod).Hours() / 24)

	for _, formation := range formations {
		quantity, _ := formation["quantity"].(float64)
		size, _ := formation["size"].(map[string]interface{})
		price, _ := size["price"].(map[string]interface{})
		cents, _ := price["cents"].(float64)

		// Calculate daily cost and multiply by the number of days in the current period
		dailyCost := (cents / 100) * quantity * (24 / 720) // Heroku bills hourly, so we divide by 720 (30 days * 24 hours)
		totalCost += dailyCost * daysInPeriod
	}

	return totalCost, nil
}

func (h *HerokuProvider) getPipelines() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/pipelines", herokuAPIRootURL)
	return h.makeRequest(url)
}

func (h *HerokuProvider) getAppPipelineCoupling(appName string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/apps/%s/pipeline-couplings", herokuAPIRootURL, appName)
	results, err := h.makeRequest(url)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil // App is not coupled to a pipeline
	}
	return results[0], nil
}

func (h *HerokuProvider) getPipelineReviewApps(pipelineID string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/pipelines/%s/review-apps", herokuAPIRootURL, pipelineID)
	return h.makeRequest(url)
}

func (h *HerokuProvider) getPipelineReviewAppConfig(pipelineID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/pipelines/%s/review-app-config", herokuAPIRootURL, pipelineID)
	results, err := h.makeRequest(url)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (h *HerokuProvider) makeRequest(url string) ([]map[string]interface{}, error) {
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var herokuErr HerokuError
		if err := json.Unmarshal(body, &herokuErr); err == nil && herokuErr.ID == "not_found" {
			return []map[string]interface{}{}, nil // Return empty list for "not found" cases
		}
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result []map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		// If unmarshaling to slice fails, try unmarshaling to single object
		var singleResult map[string]interface{}
		if err := json.Unmarshal(body, &singleResult); err == nil {
			result = []map[string]interface{}{singleResult}
		} else {
			return nil, fmt.Errorf("error decoding response: %w", err)
		}
	}

	return result, nil
}

func (h *HerokuProvider) makeRequestConfig(url string) (map[string]string, error) {
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var herokuErr HerokuError
		if err := json.Unmarshal(body, &herokuErr); err == nil && herokuErr.ID == "not_found" {
			return map[string]string{}, nil // Return empty map for "not found" cases
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
