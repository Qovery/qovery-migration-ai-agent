package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/bedrock"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/sources"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "embed"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/qovery"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
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

func GenerateHerokuMigrationAssets(herokuAPIKey, awsKey, awsSecret, awsRegion, qoveryAPIKey, githubToken, destination string, progressChan chan<- ProgressUpdate) (*Assets, error) {
	progressChan <- ProgressUpdate{Stage: "Fetching configs", Progress: 0.1}

	herokuProvider := sources.NewHerokuProvider(herokuAPIKey)
	configs, err := herokuProvider.GetAllAppsConfig()
	if err != nil {
		return nil, fmt.Errorf("error fetching Heroku configs: %w", err)
	}

	return GenerateMigrationAssets(configs, awsKey, awsSecret, awsRegion, qoveryAPIKey, githubToken, destination, progressChan)
}

func GenerateCleverCloudMigrationAssets(authToken, awsKey, awsSecret, awsRegion, qoveryAPIKey, githubToken, destination string, progressChan chan<- ProgressUpdate) (*Assets, error) {
	progressChan <- ProgressUpdate{Stage: "Fetching configs", Progress: 0.1}

	clevercloudProvider := sources.NewCleverCloudProvider(authToken)
	configs, err := clevercloudProvider.GetAllAppsConfig()
	if err != nil {
		return nil, fmt.Errorf("error fetching Clever Cloud configs: %w", err)
	}

	return GenerateMigrationAssets(configs, awsKey, awsSecret, awsRegion, qoveryAPIKey, githubToken, destination, progressChan)
}

// GenerateMigrationAssets generates all necessary assets for migration and reports progress
func GenerateMigrationAssets(configs []sources.AppConfig, awsKey, awsSecret, awsRegion, qoveryAPIKey, githubToken, destination string, progressChan chan<- ProgressUpdate) (*Assets, error) {
	// Initialize Bedrock client with AWS credentials
	bedrockClient := bedrock.NewBedrockClient(awsKey, awsSecret, bedrock.ClientConfig{
		MaxRequestsPerMinute: 50,
		MaxRetries:           20,
		InitialRetryDelay:    1 * time.Second,
		MaxRetryDelay:        3 * time.Minute,
		AWSRegion:            awsRegion,
	})

	qoveryProvider := qovery.NewQoveryProvider(qoveryAPIKey)

	var err error

	progressChan <- ProgressUpdate{Stage: "Processing configs", Progress: 0.3}

	var qoveryConfigs = make(map[string]interface{})
	var dockerfiles []Dockerfile

	var currentCost = 0.0

	totalApps := len(configs)
	for i, app := range configs {
		appName := app.Name()
		qoveryConfig := qoveryProvider.TranslateConfig(appName, app.Map(), destination)
		qoveryConfigs[appName] = qoveryConfig
		currentCost += app.Cost()

		dockerfile, _, err := generateDockerfile(app.App(), bedrockClient)

		if err != nil {
			return nil, fmt.Errorf("error generating Dockerfile for %s: %w", appName, err)
		}

		dockerfiles = append(dockerfiles, Dockerfile{
			AppName:           appName,
			DockerfileContent: dockerfile,
		})

		progress := 0.3 + (float64(i+1) / float64(totalApps) * 0.4)
		progressChan <- ProgressUpdate{Stage: fmt.Sprintf("Processing app %d/%d", i+1, totalApps), Progress: progress}
	}

	progressChan <- ProgressUpdate{Stage: "Generating Terraform configs", Progress: 0.7}

	// TODO export loadQoveryTerraformDocMarkdown as a parameter
	generatedTerraformFiles, err := generateTerraformFiles(qoveryConfigs, destination, bedrockClient, githubToken, false)
	// check there is no error in the generation of Terraform files
	for _, generatedTerraformFile := range generatedTerraformFiles {
		if generatedTerraformFile.MainTf == "" {
			return nil, fmt.Errorf("error generating Terraform configs for %s: %w", generatedTerraformFile.AppName, err)
		}
	}

	progressChan <- ProgressUpdate{Stage: "Estimating costs", Progress: 0.9}

	// TODO fix the cost estimation
	// costEstimation, costEstimationPrompt, err := EstimateWorkloadCosts(terraformMain, currentCost, claudeClient)
	// prompts = append(prompts, AssetsPrompt{Type: "CostEstimation", Name: "cost_estimation_report.md", Prompt: costEstimationPrompt, Result: costEstimation})

	if err != nil {
		return nil, fmt.Errorf("error estimating workload costs: %w", err)
	}

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

			prompt := fmt.Sprintf(`CONTEXT:
This function must return the Terraform configuration main.tf and variables.tf that will be used to deploy the application with Qovery.

OUTPUT FORMAT REQUIREMENTS:
Provide two separate configurations
1. A main.tf file containing the Terraform configuration for the app.
2. A variables.tf file containing the variables for the main.tf file.
Format the response as a tuple of two strings with a "|||" separator: (main_tf_content|||variables_tf_content) without anything else. No introduction our final sentences because I will parse the output.

Example of the output:
(terraform {
  required_providers {
    qovery = {
      source = "qovery/qovery"
    }
  }
}

provider "qovery" {
  token = var.qovery_access_token
}|||variable "qovery_access_token" {
  type        = string
  description = "Qovery API access token"
}

variable "environment_id" {
  type        = string
  description = "Qovery environment ID"
})

So my parser function in Golang can parse the output and extract the two strings.

GENERATE A CONSOLIDATED TERRAFORM CONFIGURATION FOR QOVERY THAT INCLUDE THE FOLLOWING APP AND THE DEPENDENCIES (DATABASES, SERVICES, ETC):
%s

TERRAFORM GENERATION INSTRUCTIONS:
- Don't use Buildpacks, only use Dockerfiles for build_mode.
- Export secrets or sensitive information (E.g environment variable key with name containaing SECRET, KEY, URI, TOKEN, and every value that looks like a secret) from the main.tf file into the variables.tf with no default value.
- If an application refer to a database that is created by another application, make sure to use the same existing database in the Terraform configuration.
- If an application to another application via the environment variables, make sure to use the "environment_variable_aliases" from the Qovery Terraform Provider resource (if available. cf doc).
- If in the service you see an application that can be provided by a container image from the DockerHub, use the "container_image" from the Qovery Terraform Provider resource (if available. cf doc).
- If the configuration has different pipelines/stages/environments, make sure to create different Qovery environments for each set of services/applications/databases.
- If some services use the "review app" then turn on the preview environment for them with Qoverys's Terraform Provider.
- The cluster and environment resources are not required in the configuration. Export the cluster and environment ids as variables.
- Include comment into the Terraform files to explain the configuration if needed - users are technical but can be not familiar with Terraform.
- Try to optimize the Terraform configuration as much as possible.
- Don't include Terraform resources like qovery_deployment, qovery_project, qovery_environment, and qovery_cluster - use the variable references to them instead.
- Refer to the Qovery Terraform Provider Documentation below to see all the options of the provider and how to use it:
%s

USE THE FOLLOWING TERRAFORM EXAMPLES AS REFERENCE TO GENERATE THE CONFIGURATION:
%s`, string(qoveryConfigValueJSON), string(qoveryTerraformDocMarkdownJSON), string(examplesJSON))

			response, err := bedrockClient.Messages(prompt)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
						Prompt:  prompt,
					},
					err: fmt.Errorf("error generating Terraform configs for %s: %w", appName, err),
				}
				return
			}

			mainTf, variablesTf, err := parseTerraformResponse(response)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
						Prompt:  prompt,
					},
					err: fmt.Errorf("error parsing Terraform response for %s: %w", appName, err),
				}
				return
			}

			// Validate the Terraform configuration
			finalMainTf, err := validateTerraform(mainTf, variablesTf, bedrockClient)
			if err != nil {
				resultChan <- result{
					terraform: GeneratedTerraform{
						AppName: appName,
						Prompt:  prompt,
					},
					err: fmt.Errorf("error validating Terraform configuration for %s: %w", appName, err),
				}
				return
			}

			resultChan <- result{
				terraform: GeneratedTerraform{
					AppName:     appName,
					MainTf:      finalMainTf,
					VariablesTf: variablesTf,
					Prompt:      prompt,
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

	// If there were any errors, return the first one
	if len(errors) > 0 {
		return generatedTerraformFiles, fmt.Errorf("errors occurred during parallel processing: %v", errors[0])
	}

	return generatedTerraformFiles, nil
}

// EstimateWorkloadCosts estimates the costs of running the workload and provides a comparison report
func EstimateWorkloadCosts(mainTfContent string, currentCosts float64, bedrockClient *bedrock.BedrockClient) (string, string, error) {
	// Prepare the prompt for Claude
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

	// Get Claude's response
	response, err := bedrockClient.Messages(prompt)
	if err != nil {
		return "", prompt, fmt.Errorf("error getting response from Claude: %w", err)
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

// parseTerraformResponse parses the Claude AI response for Terraform configurations
func parseTerraformResponse(response string) (string, string, error) {
	response = strings.TrimSpace(response)

	// Find the content between the first '(' and last ')'
	startIdx := strings.Index(response, "(")
	endIdx := strings.LastIndex(response, ")")

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return "", "", fmt.Errorf("could not find matching parentheses in response")
	}

	// Extract content between parentheses
	content := response[startIdx+1 : endIdx]

	// Split the content by the delimiter
	parts := strings.SplitN(content, "|||", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid response format: missing delimiter '|||'")
	}

	// Clean up each part
	mainTf := strings.TrimSpace(parts[0])
	variablesTf := strings.TrimSpace(parts[1])

	// Remove any quotes that might be present
	mainTf = strings.Trim(mainTf, "\"")
	variablesTf = strings.Trim(variablesTf, "\"")

	return mainTf, variablesTf, nil
}

// ValidateTerraform takes an original Terraform manifest, validates it, and returns the final valid manifest or an error
func validateTerraform(originalMainManifest string, originalVariablesManifest string, bedrockClient *bedrock.BedrockClient) (string, error) {
	// Create a temporary directory for Terraform files
	tempDir, err := ioutil.TempDir("", "terraform-validate")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the original manifest to a file in the temp directory
	tfFilePath := filepath.Join(tempDir, "main.tf")
	if err := ioutil.WriteFile(tfFilePath, []byte(originalMainManifest), 0644); err != nil {
		return "", fmt.Errorf("failed to write Terraform file: %w", err)
	}

	tfVarFilePath := filepath.Join(tempDir, "variables.tf")
	if err := ioutil.WriteFile(tfVarFilePath, []byte(originalVariablesManifest), 0644); err != nil {
		return "", fmt.Errorf("failed to write Terraform file: %w", err)
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
				return "", fmt.Errorf("error reading main Terraform file: %w", err)
			}

			varsContent, err := ioutil.ReadFile(tfVarFilePath)
			if err != nil {
				return "", fmt.Errorf("error reading variables Terraform file: %w", err)
			}

			// Prepare the prompt for Claude to fix init errors
			prompt := fmt.Sprintf(`The following Terraform configuration failed during initialization:

Main Terraform file (main.tf):
%s

Variables file (variables.tf):
%s

The initialization error is:
%s

Please fix the Terraform configuration to resolve these initialization errors. Focus on issues like:
- Missing or incorrect provider configurations
- Invalid backend configurations
- Module source problems
- Version constraints

Provide only the corrected Terraform code for both files without any explanations. Start the main.tf content with ### MAIN.TF ### and the variables.tf content with ### VARIABLES.TF ###`, mainContent, varsContent, initOutput)

			// Delay before retrying
			time.Sleep(3 * time.Second)

			// Get Claude's response
			correctedConfig, err := bedrockClient.Messages(prompt)
			if err != nil {
				return "", fmt.Errorf("error getting response from Claude: %w", err)
			}

			// Split the response into main.tf and variables.tf content
			parts := strings.Split(correctedConfig, "### VARIABLES.TF ###")
			if len(parts) != 2 {
				return "", fmt.Errorf("invalid response format from Claude")
			}

			mainContent = []byte(strings.TrimPrefix(parts[0], "### MAIN.TF ###"))
			varsContent = []byte(parts[1])

			// Write the corrected Terraform files
			if err := ioutil.WriteFile(tfFilePath, []byte(mainContent), 0644); err != nil {
				return "", fmt.Errorf("error writing corrected main.tf: %w", err)
			}

			if err := ioutil.WriteFile(tfVarFilePath, []byte(varsContent), 0644); err != nil {
				return "", fmt.Errorf("error writing corrected variables.tf: %w", err)
			}

			fmt.Println("Applied initialization corrections from Claude. Retrying...")
			continue
		}

		// Run terraform validate
		validateCmd := exec.Command("terraform", "validate", "-json")
		validateCmd.Dir = tempDir

		output, err := validateCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Terraform validation failed: %s\n", output)

			// Read the current Terraform file
			tfContent, err := ioutil.ReadFile(tfFilePath)
			if err != nil {
				return "", fmt.Errorf("error reading Terraform file: %w", err)
			}

			// Prepare the prompt for Claude
			prompt := fmt.Sprintf(`The following Terraform configuration has validation errors:

%s

The validation error is:
%s

Please fix the Terraform configuration to resolve these errors. Provide only the corrected Terraform code without any explanations.`, tfContent, output)

			// Delay before retrying
			time.Sleep(3 * time.Second)

			// Get Claude's response
			correctedTF, err := bedrockClient.Messages(prompt)
			if err != nil {
				return "", fmt.Errorf("error getting response from Claude: %w", err)
			}

			// Write the corrected Terraform to file
			err = ioutil.WriteFile(tfFilePath, []byte(correctedTF), 0644)
			if err != nil {
				return "", fmt.Errorf("error writing corrected Terraform: %w", err)
			}

			fmt.Println("Applied validation corrections from Claude. Retrying...")
		} else {
			// Read and return the final valid Terraform manifest
			finalManifest, err := ioutil.ReadFile(tfFilePath)
			if err != nil {
				return "", fmt.Errorf("error reading final Terraform manifest: %w", err)
			}
			return string(finalManifest), nil
		}
	}

	return "", fmt.Errorf("exceeded maximum iterations (%d) without achieving a valid Terraform configuration", maxIterations)
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

		// Write main.tf
		if err := writeToFile(filepath.Join(appDir, "main.tf"), generatedTf.MainTf); err != nil {
			return fmt.Errorf("error writing main.tf: %w", err)
		}

		// Write variables.tf
		if err := writeToFile(filepath.Join(appDir, "variables.tf"), generatedTf.VariablesTf); err != nil {
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
