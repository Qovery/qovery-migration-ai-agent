package migration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClaudeClient is a mock for the ClaudeClient
type MockClaudeClient struct {
	mock.Mock
}

func (m *MockClaudeClient) Chat(prompt string) (string, error) {
	args := m.Called(prompt)
	return args.String(0), args.Error(1)
}

// MockHerokuProvider is a mock for the HerokuProvider
type MockHerokuProvider struct {
	mock.Mock
}

func (m *MockHerokuProvider) GetAllAppsConfig() ([]map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func TestGenerateMigrationAssets(t *testing.T) {
	// Create mock objects
	mockClaudeClient := new(MockClaudeClient)
	mockHerokuProvider := new(MockHerokuProvider)

	// Set up expectations
	mockHerokuProvider.On("GetAllAppsConfig").Return([]map[string]interface{}{
		{"name": "app1", "stack": "heroku-18"},
		{"name": "app2", "stack": "heroku-20"},
	}, nil)

	mockClaudeClient.On("Messages", mock.AnythingOfType("string")).Return("FROM ruby:2.7\nCMD [\"ruby\", \"app.rb\"]", nil).Times(2)
	mockClaudeClient.On("Messages", mock.AnythingOfType("string")).Return("(resource \"qovery_project\" \"my_project\" {}, variable \"qovery_api_token\" {})", nil).Once()

	// Create a test instance of GenerateMigrationAssets with mock dependencies
	testGenerateMigrationAssets := func(source, herokuAPIKey, claudeAPIKey, qoveryAPIKey, destination string) (*Assets, error) {
		herokuProvider := mockHerokuProvider
		claudeClient := mockClaudeClient
		qoveryProvider := qovery.NewQoveryProvider(qoveryAPIKey)

		var configs []map[string]interface{}
		var err error

		switch source {
		case "heroku":
			configs, err = herokuProvider.GetAllAppsConfig()
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported source: %s", source)
		}

		// ... (rest of the function remains the same as in the original GenerateMigrationAssets)
	}

	// Run the test
	assets, err := testGenerateMigrationAssets("heroku", "fake-heroku-key", "fake-claude-key", "fake-qovery-key", "aws")

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, assets)
	assert.Len(t, assets.Dockerfiles, 2)
	assert.Equal(t, "resource \"qovery_project\" \"my_project\" {}", assets.TerraformMain)
	assert.Equal(t, "variable \"qovery_api_token\" {}", assets.TerraformVariables)

	// Verify that all expectations were met
	mockHerokuProvider.AssertExpectations(t)
	mockClaudeClient.AssertExpectations(t)
}
