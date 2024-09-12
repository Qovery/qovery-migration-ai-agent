package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v39/github"
	"path/filepath"
	"qovery-ai-migration/pkg/claude"
	"qovery-ai-migration/pkg/heroku"
	"qovery-ai-migration/pkg/qovery"
	"strings"
)

// Assets represents the generated assets for migration
type Assets struct {
	TerraformMain      string
	TerraformVariables string
	Dockerfiles        []Dockerfile
}

// Dockerfile represents a generated Dockerfile for an app
type Dockerfile struct {
	AppName           string
	DockerfileContent string
}

// GenerateMigrationAssets generates all necessary assets for migration
func GenerateMigrationAssets(source, herokuAPIKey, claudeAPIKey, qoveryAPIKey, destination string) (*Assets, error) {
	claudeClient := claude.NewClaudeClient(claudeAPIKey)
	qoveryProvider := qovery.NewQoveryProvider(qoveryAPIKey)

	var configs []map[string]interface{}
	var err error

	switch source {
	case "heroku":
		herokuProvider := heroku.NewHerokuProvider(herokuAPIKey)
		configs, err = herokuProvider.GetAllAppsConfig()
		if err != nil {
			return nil, fmt.Errorf("error fetching Heroku configs: %w", err)
		}
	// Add cases for other sources here in the future
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}

	var qoveryConfigs = make(map[string]interface{})
	var dockerfiles []Dockerfile

	for _, app := range configs {
		appName := app["name"].(string)
		qoveryConfig := qoveryProvider.TranslateConfig(app, destination)
		qoveryConfigs[appName] = qoveryConfig

		dockerfile, err := generateDockerfile(app, claudeClient)
		if err != nil {
			return nil, fmt.Errorf("error generating Dockerfile for %s: %w", appName, err)
		}

		dockerfiles = append(dockerfiles, Dockerfile{
			AppName:           appName,
			DockerfileContent: dockerfile,
		})
	}

	terraformMain, terraformVariables, err := generateTerraform(qoveryConfigs, destination, claudeClient)
	if err != nil {
		return nil, fmt.Errorf("error generating Terraform configs: %w", err)
	}

	return &Assets{
		TerraformMain:      terraformMain,
		TerraformVariables: terraformVariables,
		Dockerfiles:        dockerfiles,
	}, nil
}

// generateDockerfile generates a Dockerfile for a given app configuration
func generateDockerfile(appConfig map[string]interface{}, claudeClient *claude.ClaudeClient) (string, error) {
	configJSON, err := json.Marshal(appConfig)
	if err != nil {
		return "", fmt.Errorf("error marshaling app config: %w", err)
	}

	prompt := fmt.Sprintf(`Generate a Dockerfile for the following app configuration:\n%s\n\n

Instructions:
- You should find all the information you need like the language, the potential framework and the version associated.
- The Dockerfile should be optimized for the best performance and security.
- Generate just the Dockerfile content and nothing else.
`,
		string(configJSON))
	return claudeClient.Messages(prompt)
}

// generateTerraform generates Terraform configurations for Qovery

const (
	owner      = "Qovery"
	repo       = "terraform-examples"
	exampleDir = "examples"
)

type TerraformExample struct {
	Name    string
	Content string
}

func generateTerraform(qoveryConfigs map[string]interface{}, destination string, claudeClient *claude.ClaudeClient) (string, string, error) {
	configJSON, err := json.Marshal(qoveryConfigs)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling Qovery configs: %w", err)
	}

	examples, err := loadTerraformExamples()
	if err != nil {
		return "", "", fmt.Errorf("error loading Terraform examples: %w", err)
	}

	examplesJSON, err := json.Marshal(examples)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling Terraform examples: %w", err)
	}

	prompt := fmt.Sprintf(`Generate a consolidated Terraform configuration for Qovery that includes all of the following apps:
%s
The configuration should be for the %s cloud provider.
Use the following Terraform examples as reference:
%s
Provide two separate configurations:
1. A main.tf file containing the full Terraform configuration for all apps.
2. A variables.tf file containing the Qovery API token and the necessary credentials for the %s cloud provider.
Format the response as a tuple of two strings: (main_tf_content, variables_tf_content).
Do not include anything else.`,
		string(configJSON), destination, string(examplesJSON), destination)

	response, err := claudeClient.Messages(prompt)
	if err != nil {
		return "", "", fmt.Errorf("error generating Terraform configs: %w", err)
	}

	return parseTerraformResponse(response)
}

func loadTerraformExamples() ([]TerraformExample, error) {
	client := github.NewClient(nil)

	_, dirContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, exampleDir, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching repository contents: %w", err)
	}

	var examples []TerraformExample

	for _, content := range dirContent {
		if content.GetType() == "dir" {
			mainTFPath := filepath.Join(exampleDir, content.GetName(), "main.tf")
			fileContent, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, mainTFPath, nil)
			if err != nil {
				continue // Skip if main.tf doesn't exist
			}

			decodedContent, err := fileContent.GetContent()
			if err != nil {
				return nil, fmt.Errorf("error decoding file content: %w", err)
			}

			examples = append(examples, TerraformExample{
				Name:    content.GetName(),
				Content: decodedContent,
			})
		}
	}

	return examples, nil
}

// parseTerraformResponse parses the Claude AI response for Terraform configurations
func parseTerraformResponse(response string) (string, string, error) {
	response = strings.TrimSpace(response)
	if !strings.HasPrefix(response, "(") || !strings.HasSuffix(response, ")") {
		return "", "", fmt.Errorf("invalid response format")
	}

	content := response[1 : len(response)-1]
	parts := strings.SplitN(content, ",", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid response format")
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}
