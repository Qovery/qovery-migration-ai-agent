package sources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

const (
	cleverCloudAPIRootURL = "https://api.clever-cloud.com/v2"
	cleverCloudAPIV4URL   = "https://api.clever-cloud.com/v4"
)

type CleverCloudProvider struct {
	client    *http.Client
	authToken string
}

type CleverCloudSummary struct {
	User          interface{}               `json:"user"`
	Organisations []CleverCloudOrganisation `json:"organisations"`
}

type CleverCloudOrganisation struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Applications []CleverCloudAppBasic   `json:"applications"`
	Addons       []CleverCloudAddonBasic `json:"addons"`
}

type CleverCloudAppBasic struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CleverCloudAddonBasic struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CleverCloudAppConfig struct {
	ID            string                  `json:"id"`
	MName         string                  `json:"name"`
	Description   string                  `json:"description"`
	Zone          string                  `json:"zone"`
	Instance      map[string]interface{}  `json:"instance"`
	Deployment    map[string]interface{}  `json:"deployment"`
	Vhosts        []map[string]string     `json:"vhosts"`
	CreationDate  int64                   `json:"creationDate"`
	State         string                  `json:"state"`
	Env           []map[string]string     `json:"env"`
	Addons        []CleverCloudAddonBasic `json:"addons"`
	CustomDomains []map[string]string     `json:"customDomains"`
}

type CleverCloudAddonConfig struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	RealID       string                 `json:"realId"`
	Region       string                 `json:"region"`
	Provider     map[string]interface{} `json:"provider"`
	Plan         map[string]interface{} `json:"plan"`
	CreationDate int64                  `json:"creationDate"`
	ConfigKeys   []string               `json:"configKeys"`
	EnvVars      map[string]string      `json:"envVars"`
}

func (c CleverCloudAppConfig) App() map[string]interface{} {
	return map[string]interface{}{
		"id":            c.ID,
		"name":          c.Name,
		"description":   c.Description,
		"zone":          c.Zone,
		"instance":      c.Instance,
		"deployment":    c.Deployment,
		"vhosts":        c.Vhosts,
		"creationDate":  c.CreationDate,
		"state":         c.State,
		"env":           c.Env,
		"addons":        c.Addons,
		"customDomains": c.CustomDomains,
	}
}

func (c CleverCloudAppConfig) Name() string {
	return c.MName
}

func (c CleverCloudAppConfig) Cost() float64 {
	// TODO check if there's a cost API for Clever Cloud
	return 0
}

func (c CleverCloudAppConfig) Map() map[string]interface{} {
	return map[string]interface{}{
		"app":  c.App(),
		"cost": c.Cost(),
	}
}

func NewCleverCloudProvider(authToken string) *CleverCloudProvider {
	return &CleverCloudProvider{
		client:    &http.Client{},
		authToken: authToken,
	}
}

func (c *CleverCloudProvider) GetAllAppsConfig() ([]AppConfig, error) {
	summary, err := c.getSummary()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allApps []CleverCloudAppConfig

	for _, org := range summary.Organisations {
		for _, app := range org.Applications {
			wg.Add(1)
			go func(orgID, appID string) {
				defer wg.Done()
				appConfig, err := c.getAppDetails(orgID, appID)
				if err != nil {
					fmt.Printf("Error fetching details for app %s: %v\n", appID, err)
					return
				}

				envVars, err := c.getAppEnvVars(orgID, appID)
				if err != nil {
					fmt.Printf("Error fetching env vars for app %s: %v\n", appID, err)
				} else {
					appConfig.Env = envVars
				}

				customDomains, err := c.getAppCustomDomains(orgID, appID)
				if err != nil {
					fmt.Printf("Error fetching custom domains for app %s: %v\n", appID, err)
				} else {
					appConfig.CustomDomains = customDomains
				}

				addons, err := c.getAppAddons(orgID, appID)
				if err != nil {
					fmt.Printf("Error fetching addons for app %s: %v\n", appID, err)
				} else {
					appConfig.Addons = addons
				}

				mu.Lock()
				allApps = append(allApps, appConfig)
				mu.Unlock()
			}(org.ID, app.ID)
		}
	}

	wg.Wait()

	var appConfigs []AppConfig
	for _, app := range allApps {
		appConfigs = append(appConfigs, app)
	}

	return appConfigs, nil
}

func (c *CleverCloudProvider) GetAllAddonsConfig() ([]CleverCloudAddonConfig, error) {
	summary, err := c.getSummary()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allAddons []CleverCloudAddonConfig

	for _, org := range summary.Organisations {
		for _, addon := range org.Addons {
			wg.Add(1)
			go func(orgID, addonID string) {
				defer wg.Done()
				addonConfig, err := c.getAddonDetails(orgID, addonID)
				if err != nil {
					fmt.Printf("Error fetching details for addon %s: %v\n", addonID, err)
					return
				}

				if addonConfig.Provider["id"] == "config-provider" {
					envVars, err := c.getAddonEnvVars(addonConfig.RealID)
					if err != nil {
						fmt.Printf("Error fetching env vars for addon %s: %v\n", addonID, err)
					} else {
						addonConfig.EnvVars = envVars
					}
				}

				mu.Lock()
				allAddons = append(allAddons, addonConfig)
				mu.Unlock()
			}(org.ID, addon.ID)
		}
	}

	wg.Wait()
	return allAddons, nil
}

func (c *CleverCloudProvider) getSummary() (*CleverCloudSummary, error) {
	url := fmt.Sprintf("%s/summary", cleverCloudAPIRootURL)
	var summary CleverCloudSummary
	err := c.makeRequest("GET", url, nil, &summary)
	return &summary, err
}

func (c *CleverCloudProvider) getAppDetails(orgID, appID string) (CleverCloudAppConfig, error) {
	url := fmt.Sprintf("%s/organisations/%s/applications/%s", cleverCloudAPIRootURL, orgID, appID)
	var appConfig CleverCloudAppConfig
	err := c.makeRequest("GET", url, nil, &appConfig)
	return appConfig, err
}

func (c *CleverCloudProvider) getAppEnvVars(orgID, appID string) ([]map[string]string, error) {
	url := fmt.Sprintf("%s/organisations/%s/applications/%s/env", cleverCloudAPIRootURL, orgID, appID)
	var envVars []map[string]string
	err := c.makeRequest("GET", url, nil, &envVars)
	return envVars, err
}

func (c *CleverCloudProvider) getAppCustomDomains(orgID, appID string) ([]map[string]string, error) {
	url := fmt.Sprintf("%s/organisations/%s/applications/%s/vhosts", cleverCloudAPIRootURL, orgID, appID)
	var customDomains []map[string]string
	err := c.makeRequest("GET", url, nil, &customDomains)
	return customDomains, err
}

func (c *CleverCloudProvider) getAppAddons(orgID, appID string) ([]CleverCloudAddonBasic, error) {
	url := fmt.Sprintf("%s/organisations/%s/applications/%s/addons", cleverCloudAPIRootURL, orgID, appID)
	var addons []CleverCloudAddonBasic
	err := c.makeRequest("GET", url, nil, &addons)
	return addons, err
}

func (c *CleverCloudProvider) getAddonDetails(orgID, addonID string) (CleverCloudAddonConfig, error) {
	url := fmt.Sprintf("%s/organisations/%s/addons/%s", cleverCloudAPIRootURL, orgID, addonID)
	var addonConfig CleverCloudAddonConfig
	err := c.makeRequest("GET", url, nil, &addonConfig)
	return addonConfig, err
}

func (c *CleverCloudProvider) getAddonEnvVars(realAddonID string) (map[string]string, error) {
	url := fmt.Sprintf("%s/addon-providers/config-provider/addons/%s/env", cleverCloudAPIV4URL, realAddonID)
	var envVars map[string]string
	err := c.makeRequest("GET", url, nil, &envVars)
	return envVars, err
}

func (c *CleverCloudProvider) makeRequest(method, url string, body []byte, v interface{}) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.authToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	return nil
}
