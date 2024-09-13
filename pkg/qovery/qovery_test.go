package qovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQoveryProvider_TranslateConfig(t *testing.T) {
	provider := NewQoveryProvider("fake-api-key")

	herokuConfig := map[string]interface{}{
		"name":  "test-app",
		"stack": "heroku-20",
	}

	qoveryConfig := provider.TranslateConfig(herokuConfig, "aws")

	assert.Equal(t, "test-app", qoveryConfig["app_name"])
	assert.Equal(t, "aws", qoveryConfig["destination"])
	assert.Equal(t, "heroku-20", qoveryConfig["stack"])
}
