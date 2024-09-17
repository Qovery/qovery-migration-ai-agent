package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "embed"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/claude"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/heroku"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/qovery"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

//go:embed _readme.md
var readmeContent string

// Assets represents the generated assets for migration
type Assets struct {
	ReadmeMarkdown               string
	TerraformMain                string
	TerraformVariables           string
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

// GenerateMigrationAssets generates all necessary assets for migration and reports progress
func GenerateMigrationAssets(source, herokuAPIKey, claudeAPIKey, qoveryAPIKey, githubToken, destination string, progressChan chan<- ProgressUpdate) (*Assets, error) {
	claudeClient := claude.NewClaudeClient(claudeAPIKey)
	qoveryProvider := qovery.NewQoveryProvider(qoveryAPIKey)

	var configs []heroku.AppConfig
	var err error

	// Fetch configs
	progressChan <- ProgressUpdate{Stage: "Fetching configs", Progress: 0.1}
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

	progressChan <- ProgressUpdate{Stage: "Processing configs", Progress: 0.3}

	var qoveryConfigs = make(map[string]interface{})
	var dockerfiles []Dockerfile

	var currentCost = 0.0

	totalApps := len(configs)
	for i, app := range configs {
		appName := app.App["name"].(string)
		qoveryConfig := qoveryProvider.TranslateConfig(appName, app.Map(), destination)
		qoveryConfigs[appName] = qoveryConfig
		currentCost += app.Cost

		dockerfile, err := generateDockerfile(app.App, claudeClient)
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
	terraformMain, terraformVariables, err := generateTerraform(qoveryConfigs, destination, claudeClient, githubToken, false)
	if err != nil {
		return nil, fmt.Errorf("error generating Terraform configs: %w", err)
	}

	progressChan <- ProgressUpdate{Stage: "Estimating costs", Progress: 0.9}

	costEstimation, err := EstimateWorkloadCosts(terraformMain, currentCost, claudeClient)
	if err != nil {
		return nil, fmt.Errorf("error estimating workload costs: %w", err)
	}

	assets := &Assets{
		ReadmeMarkdown:               readmeContent,
		TerraformMain:                terraformMain,
		TerraformVariables:           terraformVariables,
		Dockerfiles:                  dockerfiles,
		CostEstimationReportMarkdown: costEstimation,
	}

	progressChan <- ProgressUpdate{Stage: "Completed", Progress: 1.0}

	return assets, nil
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
`, string(configJSON))
	return claudeClient.Messages(prompt)
}

// generateTerraform generates Terraform configurations for Qovery
func generateTerraform(qoveryConfigs map[string]interface{}, destination string, claudeClient *claude.ClaudeClient,
	githubToken string, loadQoveryTerraformDocMarkdown bool) (string, string, error) {

	configJSON, err := json.Marshal(qoveryConfigs)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling Qovery configs: %w", err)
	}

	officialExamples, err := loadTerraformExamples("Qovery", "terraform-examples", "examples", githubToken)
	if err != nil {
		return "", "", fmt.Errorf("error loading Terraform examples: %w", err)
	}

	airbyteExample, err := loadTerraformExamples("evoxmusic", "qovery-airbyte", ".", githubToken)
	if err != nil {
		return "", "", fmt.Errorf("error loading Airbyte Terraform example: %w", err)
	}

	qoveryTerraformDocMarkdown, err := loadMarkdownFiles("Qovery", "terraform-provider-qovery", "main", githubToken)
	if err != nil {
		return "", "", fmt.Errorf("error loading Qovery Terraform Provider markdown documentation: %w", err)
	}

	examples := append(officialExamples, airbyteExample...)

	examplesJSON, err := json.Marshal(examples)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling Terraform examples: %w", err)
	}

	var qoveryTerraformDocMarkdownJSON []byte
	if loadQoveryTerraformDocMarkdown {
		qoveryTerraformDocMarkdownJSON = []byte("")
	} else {
		qoveryTerraformDocMarkdownJSON, err = json.Marshal(qoveryTerraformDocMarkdown)

		if err != nil {
			return "", "", fmt.Errorf("error marshaling Qovery Terraform Provider markdown documentation: %w", err)
		}
	}

	prompt := fmt.Sprintf(`Output instructions:
Provide two separate configurations
1. A main.tf file containing the full Terraform configuration for all apps.
2. A variables.tf file containing the Qovery API token and the necessary credentials for the %s cloud provider.
Format the response as a tuple of two strings with a "|||" separator: (main_tf_content|||variables_tf_content) without anything else. No introduction our final sentences.

Generate a consolidated Terraform configuration for Qovery that includes all of the following apps:
%s

The configuration should be for the %s cloud provider.
Use the following Terraform examples as reference:
%s

Additional instructions:
- Don't use Buildpacks, only use Dockerfiles for build_mode.
- Export secrets or sensitive information from the main.tf file into the variables.tf with no default value.
- If an application refer to a database that is created by another application, make sure to use the same existing database in the Terraform configuration.
- If an application to another application via the environment variables, make sure to use the "environment_variable_aliases" from the Qovery Terraform Provider resource (if available. cf doc).
- If in the service you see an application that can be provided by a container image from the DockerHub, use the "container_image" from the Qovery Terraform Provider resource (if available. cf doc).
- If the configuration has different pipelines/stages/environments, make sure to create different Qovery environments for each set of services/applications/databases.
- If some services use the "review app" then turn on the preview environment for them with Qoverys's Terraform Provider.
- Include comment into the Terraform files to explain the configuration if needed - users are technical but can be not familiar with Terraform.
- Try to optimize the Terraform configuration as much as possible.
- Refer to the Qovery Terraform Provider Documentation below to see all the options of the provider and how to use it:
%s`, destination, string(configJSON), destination, string(examplesJSON), string(qoveryTerraformDocMarkdownJSON))

	response, err := claudeClient.Messages(prompt)
	if err != nil {
		return "", "", fmt.Errorf("error generating Terraform configs: %w", err)
	}

	mainTf, variablesTf, err := parseTerraformResponse(response)

	if err != nil {
		return "", "", err
	}

	// Validate the Terraform configuration
	finalMainTf, err := validateTerraform(mainTf, variablesTf, claudeClient)
	if err != nil {
		return "", "", fmt.Errorf("error validating Terraform configuration: %w", err)
	}

	return finalMainTf, variablesTf, nil
}

// EstimateWorkloadCosts estimates the costs of running the workload and provides a comparison report
func EstimateWorkloadCosts(mainTfContent string, currentCosts float64, claudeClient *claude.ClaudeClient) (string, error) {
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
	response, err := claudeClient.Messages(prompt)
	if err != nil {
		return "", fmt.Errorf("error getting response from Claude: %w", err)
	}

	// Append contact information
	response += "\n\nFor more information or if you have any questions about migrating from Heroku to Qovery, please contact us at hello@qovery.com."

	return response, nil
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
	if !strings.HasPrefix(response, "(") || !strings.HasSuffix(response, ")") {
		return "", "", fmt.Errorf("invalid response format")
	}

	content := response[1 : len(response)-1]

	parts := strings.SplitN(content, "|||", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("ParseTerraformResponse invalid response format")
	}

	// trim leading and trailing whitespace
	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])

	// Remove the leading and trailing quotes
	parts[0] = strings.Trim(parts[0], "\"")
	parts[1] = strings.Trim(parts[1], "\"")

	return parts[0], parts[1], nil
}

// ValidateTerraform takes an original Terraform manifest, validates it, and returns the final valid manifest or an error
func validateTerraform(originalMainManifest string, originalVariablesManifest string, claudeClient *claude.ClaudeClient) (string, error) {
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
			return "", fmt.Errorf("terraform init failed: %w", err)
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

			// Delay a bit before retrying
			time.Sleep(3 * time.Second)

			// Get Claude's response
			correctedTF, err := claudeClient.Messages(prompt)
			if err != nil {
				return "", fmt.Errorf("error getting response from Claude: %w", err)
			}

			// Write the corrected Terraform to file
			err = ioutil.WriteFile(tfFilePath, []byte(correctedTF), 0644)
			if err != nil {
				return "", fmt.Errorf("error writing corrected Terraform: %w", err)
			}

			fmt.Println("Applied corrections from Claude. Retrying validation...")
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
func WriteAssets(outputDir string, assets *Assets) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Write README file
	if err := writeToFile(filepath.Join(outputDir, "README.md"), assets.ReadmeMarkdown); err != nil {
		return fmt.Errorf("error writing README.md: %w", err)
	}

	// Write Terraform main configuration
	if err := writeToFile(filepath.Join(outputDir, "main.tf"), assets.TerraformMain); err != nil {
		return fmt.Errorf("error writing main.tf: %w", err)
	}

	// Write Terraform variables
	if err := writeToFile(filepath.Join(outputDir, "variables.tf"), assets.TerraformVariables); err != nil {
		return fmt.Errorf("error writing variables.tf: %w", err)
	}

	// Write Dockerfiles
	for _, dockerfile := range assets.Dockerfiles {
		filename := fmt.Sprintf("Dockerfile-%s", dockerfile.AppName)
		if err := writeToFile(filepath.Join(outputDir, filename), dockerfile.DockerfileContent); err != nil {
			return fmt.Errorf("error writing %s: %w", filename, err)
		}
	}

	// Write cost estimation report
	if err := writeToFile(filepath.Join(outputDir, "cost_estimation_report.md"), assets.CostEstimationReportMarkdown); err != nil {
		return fmt.Errorf("error writing cost_estimation_report.md: %w", err)
	}

	return nil
}

func writeToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
