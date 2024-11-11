package migration

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/bedrock"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/qovery"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/sources"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed _readme.md
var readmeContent string

// Assets represents the generated assets for migration
type Assets struct {
	ReadmeMarkdown               string
	GeneratedTerraformFiles      []GeneratedTerraform
	Dockerfiles                  []Dockerfile
	CostEstimationReportMarkdown string
}

// Dockerfile represents a generated Dockerfile for an app
type Dockerfile struct {
	AppName           string
	DockerfileContent string
}

// ProgressUpdate represents a progress update
type ProgressUpdate struct {
	Stage    string
	Progress float64
}

func GenerateHerokuMigrationAssets(herokuAPIKey, awsKey, awsSecret, qoveryAPIKey, githubToken, destination string, bedrockClientConfig bedrock.ClientConfig, progressChan chan<- ProgressUpdate) (*Assets, error) {
	progressChan <- ProgressUpdate{Stage: "Fetching configs", Progress: 0.1}

	herokuProvider := sources.NewHerokuProvider(herokuAPIKey)
	configs, err := herokuProvider.GetAllAppsConfig()
	if err != nil {
		return nil, fmt.Errorf("error fetching Heroku configs: %w", err)
	}

	return GenerateMigrationAssets(configs, awsKey, awsSecret, qoveryAPIKey, githubToken, destination, bedrockClientConfig, progressChan)
}

func GenerateCleverCloudMigrationAssets(authToken, awsKey, awsSecret, qoveryAPIKey, githubToken, destination string, bedrockClientConfig bedrock.ClientConfig, progressChan chan<- ProgressUpdate) (*Assets, error) {
	progressChan <- ProgressUpdate{Stage: "Fetching configs", Progress: 0.1}

	clevercloudProvider := sources.NewCleverCloudProvider(authToken)
	configs, err := clevercloudProvider.GetAllAppsConfig()
	if err != nil {
		return nil, fmt.Errorf("error fetching Clever Cloud configs: %w", err)
	}

	return GenerateMigrationAssets(configs, awsKey, awsSecret, qoveryAPIKey, githubToken, destination, bedrockClientConfig, progressChan)
}

// GenerateMigrationAssets generates all necessary assets for migration and reports progress
func GenerateMigrationAssets(configs []sources.AppConfig, awsKey, awsSecret, qoveryAPIKey, githubToken, destination string, bedrockClientConfig bedrock.ClientConfig, progressChan chan<- ProgressUpdate) (*Assets, error) {
	// Initialize Bedrock client with AWS credentials
	bedrockClient, err := bedrock.NewBedrockClient(awsKey, awsSecret, bedrockClientConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing Bedrock client: %w", err)
	}

	qoveryProvider := qovery.NewQoveryProvider(qoveryAPIKey)
	progressChan <- ProgressUpdate{Stage: "Processing configs", Progress: 0.3}

	totalApps := len(configs)

	// Create channels for collecting results and errors
	type dockerfileResult struct {
		dockerfile   Dockerfile
		qoveryConfig map[string]interface{}
		appName      string
		err          error
		index        int
	}
	resultChan := make(chan dockerfileResult, totalApps)

	// Launch goroutines for parallel Dockerfile generation
	var wg sync.WaitGroup
	for i, app := range configs {
		wg.Add(1)
		go func(app sources.AppConfig, index int) {
			defer wg.Done()

			appName := app.Name()
			qoveryConfig := qoveryProvider.TranslateConfig(appName, app.Map(), destination)

			dockerfile, _, err := generateDockerfile(app.App(), bedrockClient)
			if err != nil {
				resultChan <- dockerfileResult{
					err:   fmt.Errorf("error generating Dockerfile for %s: %w", appName, err),
					index: index,
				}
				return
			}

			resultChan <- dockerfileResult{
				dockerfile: Dockerfile{
					AppName:           appName,
					DockerfileContent: dockerfile,
				},
				qoveryConfig: map[string]interface{}{
					appName: qoveryConfig,
				},
				appName: appName,
				index:   index,
			}

			progress := 0.3 + (float64(index+1) / float64(totalApps) * 0.4)
			progressChan <- ProgressUpdate{
				Stage:    fmt.Sprintf("Processing app %d/%d", index+1, totalApps),
				Progress: progress,
			}
		}(app, i)
	}

	// Close result channel once all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results maintaining original order
	dockerfiles := make([]Dockerfile, totalApps)
	qoveryConfigs := make(map[string]interface{})

	// Collect results and build maps after all goroutines complete
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		dockerfiles[result.index] = result.dockerfile
		if result.qoveryConfig != nil {
			for k, v := range result.qoveryConfig {
				qoveryConfigs[k] = v
			}
		}
	}

	progressChan <- ProgressUpdate{Stage: "Generating Terraform configs", Progress: 0.7}

	generatedTerraformFiles, err := generateTerraformFiles(qoveryConfigs, destination, bedrockClient, githubToken, false)
	if err != nil {
		return nil, fmt.Errorf("error generating Terraform configs: %w", err)
	}

	progressChan <- ProgressUpdate{Stage: "Estimating costs", Progress: 0.9}

	assets := &Assets{
		ReadmeMarkdown:               readmeContent,
		GeneratedTerraformFiles:      generatedTerraformFiles,
		Dockerfiles:                  dockerfiles,
		CostEstimationReportMarkdown: "",
	}

	progressChan <- ProgressUpdate{Stage: "Completed", Progress: 1.0}

	return assets, nil
}

// generateDockerfile generates a Dockerfile for a given app configuration
func generateDockerfile(appConfig map[string]interface{}, bedrockClient *bedrock.BedrockClient) (string, string, error) {
	configJSON, err := json.Marshal(appConfig)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling app config: %w", err)
	}

	prompt := fmt.Sprintf(`GENERATE A DOCKERFILE FOR THE FOLLOWING APP CONFIGURATION:\n%s\n\n
INSTRUCTIONS THAT MUST BE FOLLOWED:
- You should find all the information you need like the language, the potential framework and the version associated.
- The Dockerfile should be optimized for the best performance and security.
- Generate just the Dockerfile content and nothing else.
`, string(configJSON))

	result, err := bedrockClient.Messages(prompt)
	return result, prompt, err
}

type GeneratedTerraform struct {
	AppName     string
	MainTf      string
	VariablesTf string
	Prompt      string
}

func (g GeneratedTerraform) SanitizeAppName() string {
	s := strings.ReplaceAll(g.AppName, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return strings.ToLower(s)
}

// generateTerraformFiles generates Terraform configurations for Qovery in parallel
func generateTerraformFiles(qoveryConfigs map[string]interface{}, destination string, bedrockClient *bedrock.BedrockClient,
	githubToken string, loadQoveryTerraformDocMarkdown bool) ([]GeneratedTerraform, error) {

	officialExamples, err := loadTerraformExamples("Qovery", "terraform-examples", "examples", githubToken)
	if err != nil {
		return nil, fmt.Errorf("error loading Terraform examples: %w", err)
	}

	airbyteExample, err := loadTerraformExamples("evoxmusic", "qovery-airbyte", ".", githubToken)
	if err != nil {
		return nil, fmt.Errorf("error loading Airbyte Terraform example: %w", err)
	}

	qoveryTerraformDocMarkdown, err := loadMarkdownFiles("Qovery", "terraform-provider-qovery", "main", githubToken)
	if err != nil {
		return nil, fmt.Errorf("error loading Qovery Terraform Provider markdown documentation: %w", err)
	}

	examples := append(officialExamples, airbyteExample...)

	examplesJSON, err := json.Marshal(examples)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Terraform examples: %w", err)
	}

	var qoveryTerraformDocMarkdownJSON []byte
	if loadQoveryTerraformDocMarkdown {
		qoveryTerraformDocMarkdownJSON = []byte("")
	} else {
		qoveryTerraformDocMarkdownJSON, err = json.Marshal(qoveryTerraformDocMarkdown)
		if err != nil {
			return nil, fmt.Errorf("error marshaling Qovery Terraform Provider markdown documentation: %w", err)
		}
	}

	// Create channels for results and errors
	type result struct {
		terraform GeneratedTerraform
		err       error
	}
	resultChan := make(chan result, len(qoveryConfigs))

	// Create a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Process each app configuration in parallel
	for appName, qoveryConfigValue := range qoveryConfigs {
		wg.Add(1)
		go func(appName string, qoveryConfigValue interface{}) {
			defer wg.Done()

			qoveryConfigValueJSON, err := json.Marshal(qoveryConfigValue)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
					},
					err: fmt.Errorf("error marshaling Qovery config for %s: %w", appName, err),
				}
				return
			}

			// First request: Generate main.tf
			mainTfPrompt := fmt.Sprintf(`CONTEXT:
This function must return the main.tf Terraform configuration that will be used to deploy the application with Qovery.

OUTPUT FORMAT REQUIREMENTS:
Provide only the main.tf configuration without any variables declarations.
Format the response as a single string containing only the main.tf content, without any separators or additional text. The response must look like this:
terraform {
  required_providers {
    qovery = {
      source = "qovery/qovery"
    }
  }
}

provider "qovery" {
  token = var.qovery_access_token
}

GENERATE A CONSOLIDATED TERRAFORM CONFIGURATION FOR QOVERY THAT INCLUDE THE FOLLOWING APP AND THE DEPENDENCIES (DATABASES, SERVICES, ETC):
%s

TERRAFORM GENERATION INSTRUCTIONS:
- Don't use Buildpacks, only use Dockerfiles for build_mode.
- Export secrets or sensitive information (E.g environment variable key with name containaing SECRET, KEY, URI, TOKEN, and every value that looks like a secret) from the main.tf file into variables.
- If an application refer to a database that is created by another application, make sure to use the same existing database in the Terraform configuration.
- If an application to another application via the environment variables, make sure to use the "environment_variable_aliases" from the Qovery Terraform Provider resource (if available. cf doc).
- If in the service you see an application that can be provided by a container image from the DockerHub, use the "container_image" from the Qovery Terraform Provider resource (if available. cf doc).
- If the configuration has different pipelines/stages/environments, make sure to create different Qovery environments for each set of services/applications/databases.
- If some services use the "review app" then turn on the preview environment for them with Qoverys's Terraform Provider.
- The cluster and environment resources are not required in the configuration. Export the cluster and environment ids as variables.
- Include comment into the Terraform files to explain the configuration if needed - users are technical but can be not familiar with Terraform.
- When setting up healthchecks for the services, make sure to use scheme "HTTP" for the healthcheck type http and "TCP" for the healthcheck type tcp. Refer to the Qovery Terraform Provider Documentation for more information.
- Try to optimize the Terraform configuration as much as possible.
- Don't include Qovery Terraform resources "qovery_deployment", "qovery_project", "qovery_environment", and "qovery_cluster" in the main.tf; use the variable references "var.project_id" etc.. to them instead.
- Refer to the Qovery Terraform Provider Documentation below to see all the options of the provider and how to use it:
%s

USE THE FOLLOWING TERRAFORM EXAMPLES AS REFERENCE TO GENERATE THE CONFIGURATION:
%s`, string(qoveryConfigValueJSON), string(qoveryTerraformDocMarkdownJSON), string(examplesJSON))

			mainTfResponse, err := bedrockClient.Messages(mainTfPrompt)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
						Prompt:  mainTfPrompt,
					},
					err: fmt.Errorf("error generating main.tf for %s: %w", appName, err),
				}
				return
			}

			// Parse and validate main.tf
			mainTf, err := parseMainTfResponse(mainTfResponse)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
						Prompt:  mainTfPrompt,
					},
					err: fmt.Errorf("error parsing main.tf response for %s: %w", appName, err),
				}
				return
			}

			// Second request: Generate variables.tf based on main.tf
			variablesTfPrompt := fmt.Sprintf(`CONTEXT:
Generate the variables.tf file for the following main.tf Terraform configuration:

%s

VARIABLE GENERATION INSTRUCTIONS:
- Include all necessary variables referenced in the main.tf file
- Don't provide default values for sensitive information (secrets, tokens, keys, URIs)
- Include appropriate descriptions for all variables
- Group related variables together with comments
- Include type constraints for all variables
- Make sure to include all environment IDs, cluster IDs, and other referenced variables from main.tf

OUTPUT FORMAT REQUIREMENTS:
Provide only the variables.tf configuration.
Format the response as a single string containing only the variables.tf content, without any separators or additional text. The response must look like this:
variable "project_id" {
  type        = string
  description = "The ID of the Qovery project"
}

variable "environment_id" {
  type        = string
  description = "The ID of the Qovery environment"
}

variable "application_name" {
  type        = string
  description = "The name of the application"
}`, mainTf)

			variablesTfResponse, err := bedrockClient.Messages(variablesTfPrompt)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
						MainTf:  mainTf,
						Prompt:  variablesTfPrompt,
					},
					err: fmt.Errorf("error generating variables.tf for %s: %w", appName, err),
				}
				return
			}

			// Parse variables.tf
			variablesTf, err := parseVariablesTfResponse(variablesTfResponse)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
						MainTf:  mainTf,
						Prompt:  variablesTfPrompt,
					},
					err: fmt.Errorf("error parsing variables.tf response for %s: %w", appName, err),
				}
				return
			}

			// Validate the complete Terraform configuration
			finalMainTf, finalVariablesTf, err := validateTerraform(mainTf, variablesTf, bedrockClient)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
						Prompt:  mainTfPrompt + "\n\n" + variablesTfPrompt,
					},
					err: fmt.Errorf("error validating Terraform configuration for %s: %w", appName, err),
				}
				return
			}

			resultChan <- result{
				terraform: GeneratedTerraform{
					AppName:     appName,
					MainTf:      finalMainTf,
					VariablesTf: finalVariablesTf,
					Prompt:      mainTfPrompt + "\n\n" + variablesTfPrompt,
				},
				err: nil,
			}
		}(appName, qoveryConfigValue)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and check for errors
	var generatedTerraformFiles []GeneratedTerraform
	var errors []error

	for result := range resultChan {
		if result.err != nil {
			errors = append(errors, result.err)
		}
		generatedTerraformFiles = append(generatedTerraformFiles, result.terraform)
	}

	// we even return the errors to the user - they can be useful for debugging

	return generatedTerraformFiles, nil
}

// Helper functions to parse individual responses
func parseMainTfResponse(response string) (string, error) {
	mainTf := strings.TrimSpace(response)
	if mainTf == "" {
		return "", fmt.Errorf("empty main.tf response")
	}
	return mainTf, nil
}

func parseVariablesTfResponse(response string) (string, error) {
	variablesTf := strings.TrimSpace(response)
	if variablesTf == "" {
		return "", fmt.Errorf("empty variables.tf response")
	}
	return variablesTf, nil
}

// EstimateWorkloadCosts estimates the costs of running the workload and provides a comparison report
func EstimateWorkloadCosts(mainTfContent string, currentCosts float64, bedrockClient *bedrock.BedrockClient) (string, string, error) {
	// Prepare the prompt for Bedrock
	prompt := fmt.Sprintf(`Given the following Terraform configuration for Qovery:

%s

And the following comparison information between Qovery + Cloud Provider and Heroku:

| Feature | Qovery + Cloud Provider | Heroku |
|---------|-------------------------|--------|
| Applications run in your own cloud | ✅ | ❌ |
| Private VPC | ✅ | Enterprise only |
| Autoscaling | ✅ | Performance dynos only |
| Multiple Single tenant infrastructure | ✅ | ❌ |
| SOC2 & HIPAA compliance | ✅ | Enterprise only |
| Microservices support | ✅ | ❌ |
| Mono repository support | ✅ | ❌ |
| Static IPs | ✅ | ❌ |
| Global regions availability | Many (US, EU, Asia, etc.) | Limited (US, EU, AU) |
| Cost at scale (150 instances) | $31K/year | $450K/year |

Please provide a comprehensive cost estimation report in Markdown format. Include the following:

1. Estimated monthly costs for running this workload on the cloud provider specified in the Terraform configuration.
2. A detailed breakdown of the costs for each resource.
3. Comparison with the current costs of $%.2f per month on Heroku.
4. An analysis of cost-effectiveness, considering both direct costs and potential indirect savings from improved features and flexibility.
5. Estimated costs if the workload were to be run on other major cloud providers (AWS, GCP, Azure) for comparison.
6. A summary recommendation on whether to proceed with the migration, considering both costs and feature benefits.

Also, include the following Qovery-specific costs in your calculations:
- Managed cluster: $199/month (or $0 if using a self-managed Kubernetes cluster)
- 1 user: $29/month
- Deployment: $0.16 per minute (free up to 1000 minutes per month if using their own CI/CD)

Consider potential cost savings from:
- More efficient resource utilization
- Autoscaling capabilities
- Reduced need for enterprise-level features that are standard with Qovery
- Improved developer productivity due to better tooling and flexibility

Provide a comprehensive report that a decision-maker could use to determine if migration is worthwhile, considering both immediate cost impacts and long-term strategic benefits.

Important for the report: use as much as possible tables for the comparison and make it easy to read.
`, mainTfContent, currentCosts)

	// Get Bedrock's response
	response, err := bedrockClient.Messages(prompt)
	if err != nil {
		return "", prompt, fmt.Errorf("error getting response from Bedrock: %w", err)
	}

	// Append contact information
	response += "\n\nFor more information or if you have any questions about migrating from Heroku to Qovery, please contact us at hello@qovery.com."

	return response, prompt, nil
}

// NewGitHubClient creates a new GitHub client with optional authentication
func NewGitHubClient(token string) *github.Client {
	ctx := context.Background()
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		return github.NewClient(tc)
	}
	return github.NewClient(nil)
}

func loadTerraformExamples(owner, repo, exampleDir, token string) ([]map[string]string, error) {
	client := NewGitHubClient(token)

	_, dirContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, exampleDir, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching repository contents: %w", err)
	}

	var examples []map[string]string

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

			examples = append(examples, map[string]string{
				"name":    content.GetName(),
				"content": decodedContent,
			})
		}
	}

	return examples, nil
}

func loadMarkdownFiles(owner, repo, branch, token string) (map[string]string, error) {
	client := NewGitHubClient(token)
	ctx := context.Background()

	result := make(map[string]string)

	// Folders to search
	folders := []string{"", "data-sources", "resources"}

	for _, folder := range folders {
		// List files in the folder
		_, directoryContent, _, err := client.Repositories.GetContents(ctx, owner, repo, "docs/"+folder, &github.RepositoryContentGetOptions{Ref: branch})
		if err != nil {
			return nil, fmt.Errorf("error listing contents of %s: %v", folder, err)
		}

		for _, content := range directoryContent {
			if content.GetType() == "file" && strings.HasSuffix(content.GetName(), ".md") {
				// Get file content
				fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, content.GetPath(), &github.RepositoryContentGetOptions{Ref: branch})
				if err != nil {
					return nil, fmt.Errorf("error getting content of %s: %v", content.GetPath(), err)
				}

				// Decode content
				decodedContent, err := fileContent.GetContent()
				if err != nil {
					return nil, fmt.Errorf("error decoding content of %s: %v", content.GetPath(), err)
				}

				// Add to result map
				result[content.GetPath()] = decodedContent
			}
		}
	}

	return result, nil
}

// ValidateTerraform takes an original Terraform manifest, validates it, and returns the final valid manifest or an error
func validateTerraform(originalMainManifest string, originalVariablesManifest string, bedrockClient *bedrock.BedrockClient) (string, string, error) {
	// Create a temporary directory for Terraform files
	tempDir, err := ioutil.TempDir("", "terraform-validate")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the original manifests to files in the temp directory
	tfFilePath := filepath.Join(tempDir, "main.tf")
	if err := ioutil.WriteFile(tfFilePath, []byte(originalMainManifest), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write main.tf file: %w", err)
	}

	tfVarFilePath := filepath.Join(tempDir, "variables.tf")
	if err := ioutil.WriteFile(tfVarFilePath, []byte(originalVariablesManifest), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write variables.tf file: %w", err)
	}

	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		fmt.Printf("Iteration %d:\n", i+1)

		// Run terraform init
		initCmd := exec.Command("terraform", "init")
		initCmd.Dir = tempDir
		initOutput, err := initCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Terraform init failed: %s\n", initOutput)

			// Read the current Terraform files
			mainContent, err := ioutil.ReadFile(tfFilePath)
			if err != nil {
				return "", "", fmt.Errorf("error reading main.tf file: %w", err)
			}

			varsContent, err := ioutil.ReadFile(tfVarFilePath)
			if err != nil {
				return "", "", fmt.Errorf("error reading variables.tf file: %w", err)
			}

			// First prompt for main.tf fixes, including variables.tf content
			mainPrompt := fmt.Sprintf(`The following Terraform configuration failed during initialization:

Main Terraform file (main.tf):
%s

Current variables file (variables.tf):
%s

The initialization error is:
%s

Please fix the main.tf configuration to resolve these initialization errors while ensuring compatibility with the variables.tf file. Focus on issues like:
- Missing or incorrect provider configurations
- Invalid backend configurations
- Module source problems
- Version constraints
- Variable references matching variables.tf declarations

Provide only the corrected main.tf code without any explanations and no formatting. Output must look like this:
terraform {
  required_providers {
    qovery = {
      source = "qovery/qovery"
    }
  }
}

provider "qovery" {
  token = var.qovery_access_token
}`, mainContent, varsContent, initOutput)

			// Get Bedrock's response for main.tf
			correctedMain, err := bedrockClient.Messages(mainPrompt)
			if err != nil {
				return "", "", fmt.Errorf("error getting response from Bedrock for main.tf: %w", err)
			}

			// Write the corrected main.tf first
			if err := ioutil.WriteFile(tfFilePath, []byte(correctedMain), 0644); err != nil {
				return "", "", fmt.Errorf("error writing corrected main.tf: %w", err)
			}

			// Second prompt for variables.tf fixes, including the corrected main.tf
			varsPrompt := fmt.Sprintf(`The following Terraform configuration failed during initialization:

Current main.tf (already corrected):
%s

Variables file (variables.tf):
%s

The initialization error is:
%s

Please fix the variables.tf configuration to resolve these initialization errors while ensuring compatibility with the main.tf file. Focus on issues like:
- Variable declarations matching those referenced in main.tf
- Type constraints
- Default values
- Variable validation rules

Provide only the corrected variables.tf code without any explanations and no formatting. Output must look like this:
variable "project_id" {
  type        = string
  description = "The ID of the Qovery project"
}

variable "environment_id" {
  type        = string
  description = "The ID of the Qovery environment"
}

variable "application_name" {
  type        = string
  description = "The name of the application"
}`, correctedMain, varsContent, initOutput)

			// Get Bedrock's response for variables.tf
			correctedVars, err := bedrockClient.Messages(varsPrompt)
			if err != nil {
				return "", "", fmt.Errorf("error getting response from Bedrock for variables.tf: %w", err)
			}

			if err := ioutil.WriteFile(tfVarFilePath, []byte(correctedVars), 0644); err != nil {
				return "", "", fmt.Errorf("error writing corrected variables.tf: %w", err)
			}

			fmt.Println("Applied initialization corrections from Bedrock. Retrying...")
			continue
		}

		// Run terraform validate
		validateCmd := exec.Command("terraform", "validate", "-json")
		validateCmd.Dir = tempDir

		output, err := validateCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Terraform validation failed: %s\n", output)

			// Read the current Terraform files
			mainContent, err := ioutil.ReadFile(tfFilePath)
			if err != nil {
				return "", "", fmt.Errorf("error reading main.tf file: %w", err)
			}

			varsContent, err := ioutil.ReadFile(tfVarFilePath)
			if err != nil {
				return "", "", fmt.Errorf("error reading variables.tf file: %w", err)
			}

			// Prompt for main.tf validation fixes, including variables.tf
			mainPrompt := fmt.Sprintf(`The following Terraform configuration has validation errors:

Current main.tf:
%s

Current variables.tf:
%s

The validation error is:
%s

Please fix the main.tf configuration to resolve these errors while ensuring compatibility with variables.tf. Provide only the corrected code without any explanations. Output must look like this:
terraform {
  required_providers {
    qovery = {
      source = "qovery/qovery"
    }
  }
}

provider "qovery" {
  token = var.qovery_access_token
}`, mainContent, varsContent, output)

			// Get Bedrock's response for main.tf
			correctedMain, err := bedrockClient.Messages(mainPrompt)
			if err != nil {
				return "", "", fmt.Errorf("error getting response from Bedrock for main.tf: %w", err)
			}

			// Write the corrected main.tf first
			if err := ioutil.WriteFile(tfFilePath, []byte(correctedMain), 0644); err != nil {
				return "", "", fmt.Errorf("error writing corrected main.tf: %w", err)
			}

			// Prompt for variables.tf validation fixes, including corrected main.tf
			varsPrompt := fmt.Sprintf(`The following Terraform configuration has validation errors:

Current main.tf (already corrected):
%s

Current variables.tf:
%s

The validation error is:
%s

Please fix the variables.tf configuration to resolve these errors while ensuring compatibility with main.tf. Provide only the corrected code without any explanations. Output must look like this:
variable "project_id" {
  type        = string
  description = "The ID of the Qovery project"
}

variable "environment_id" {
  type        = string
  description = "The ID of the Qovery environment"
}

variable "application_name" {
  type        = string
  description = "The name of the application"
}`, correctedMain, varsContent, output)

			// Get Bedrock's response for variables.tf
			correctedVars, err := bedrockClient.Messages(varsPrompt)
			if err != nil {
				return "", "", fmt.Errorf("error getting response from Bedrock for variables.tf: %w", err)
			}

			if err := ioutil.WriteFile(tfVarFilePath, []byte(correctedVars), 0644); err != nil {
				return "", "", fmt.Errorf("error writing corrected variables.tf: %w", err)
			}

			fmt.Println("Applied validation corrections from Bedrock. Retrying...")
		} else {
			// Read and return both final valid Terraform manifests
			finalMain, err := ioutil.ReadFile(tfFilePath)
			if err != nil {
				return "", "", fmt.Errorf("error reading final main.tf manifest: %w", err)
			}

			finalVars, err := ioutil.ReadFile(tfVarFilePath)
			if err != nil {
				return "", "", fmt.Errorf("error reading final variables.tf manifest: %w", err)
			}

			return string(finalMain), string(finalVars), nil
		}
	}

	return "", "", fmt.Errorf("exceeded maximum iterations (%d) without achieving a valid Terraform configuration", maxIterations)
}

// WriteAssets writes the generated assets to the output directory
func WriteAssets(outputDir string, assets *Assets, writePrompts bool) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Write README file
	if err := writeToFile(filepath.Join(outputDir, "README.md"), assets.ReadmeMarkdown); err != nil {
		return fmt.Errorf("error writing README.md: %w", err)
	}

	// Write Terraform files for each app in their own directory
	for _, generatedTf := range assets.GeneratedTerraformFiles {
		appDir := filepath.Join(outputDir, generatedTf.SanitizeAppName())
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return fmt.Errorf("error creating app directory: %w", err)
		}

		mainTf := generatedTf.MainTf
		if generatedTf.MainTf == "" {
			mainTf = "# An error occurred while generating the main.tf file. Please refer to the prompt for more information."
		}

		// Write main.tf
		if err := writeToFile(filepath.Join(appDir, "main.tf"), mainTf); err != nil {
			return fmt.Errorf("error writing main.tf: %w", err)
		}

		variablesTf := generatedTf.VariablesTf
		if generatedTf.VariablesTf == "" {
			variablesTf = "# An error occurred while generating the variables.tf file. Please refer to the prompt for more information."
		}

		// Write variables.tf
		if err := writeToFile(filepath.Join(appDir, "variables.tf"), variablesTf); err != nil {
			return fmt.Errorf("error writing variables.tf: %w", err)
		}

		// check if the app has a dockerfile associated
		for _, dockerfile := range assets.Dockerfiles {
			if dockerfile.AppName == generatedTf.AppName {
				// Write Dockerfile
				if err := writeToFile(filepath.Join(appDir, "Dockerfile"), dockerfile.DockerfileContent); err != nil {
					return fmt.Errorf("error writing Dockerfile: %w", err)
				}
			}
		}
	}

	// Write cost estimation report
	if err := writeToFile(filepath.Join(outputDir, "cost_estimation_report.md"), assets.CostEstimationReportMarkdown); err != nil {
		return fmt.Errorf("error writing cost_estimation_report.md: %w", err)
	}

	if writePrompts {
		// Write generated terraform files with Prompts into JSON file
		generatedTfFilesWithPromptsJSON, err := json.Marshal(assets.GeneratedTerraformFiles)
		if err != nil {
			return fmt.Errorf("error marshaling prompts: %w", err)
		}

		if err := writeToFile(filepath.Join(outputDir, "generated_tf_files_with_prompts.json"), string(generatedTfFilesWithPromptsJSON)); err != nil {
			return fmt.Errorf("error writing generated_tf_files_with_prompts.json: %w", err)
		}

		// Write Dockerfiles with Prompts into JSON file
		dockerfilesWithPromptsJSON, err := json.Marshal(assets.Dockerfiles)
		if err != nil {
			return fmt.Errorf("error marshaling prompts: %w", err)
		}

		if err := writeToFile(filepath.Join(outputDir, "dockerfiles_with_prompts.json"), string(dockerfilesWithPromptsJSON)); err != nil {
			return fmt.Errorf("error writing dockerfiles_with_prompts.json: %w", err)
		}

		// TODO cost estimation prompts
	}

	return nil
}

func writeToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
