package qovery

// QoveryProvider represents a client for interacting with Qovery
type QoveryProvider struct {
	// for later use to fetch the configuration from current qovery account
	APIKey string
}

// NewQoveryProvider creates a new QoveryProvider with the given API key
func NewQoveryProvider(apiKey string) *QoveryProvider {
	return &QoveryProvider{APIKey: apiKey}
}

// TranslateConfig translates a PaaS configuration to a Qovery configuration
func (q *QoveryProvider) TranslateConfig(appName string, configMap map[string]interface{}, destination string) map[string]interface{} {
	return map[string]interface{}{
		"app_name":    appName,
		"destination": destination,
		"stack":       configMap,
	}
}
