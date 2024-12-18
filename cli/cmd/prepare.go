package cmd

import (
	"fmt"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/bedrock"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/migration"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"os"
)

var (
	source      string
	destination string
	outputDir   string
)

// prepareCmd represents the prepare command
var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare migration assets for Qovery",
	Long:  `This command prepares migration assets for Qovery, generating necessary Terraform configurations and Dockerfiles.`,
	Run:   runPrepare,
}

func init() {
	rootCmd.AddCommand(prepareCmd)
	prepareCmd.Flags().StringVarP(&source, "from", "f", "", "Source platform (e.g., 'heroku', 'clevercloud') (required)")
	prepareCmd.Flags().StringVarP(&destination, "to", "t", "", "Destination cloud provider (aws, gcp, or scaleway) (required)")
	prepareCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for generated files")
	_ = prepareCmd.MarkFlagRequired("from")
	_ = prepareCmd.MarkFlagRequired("to")
}

func runPrepare(cmd *cobra.Command, args []string) {
	// Validate source
	if source != "heroku" && source != "clevercloud" {
		fmt.Println("Error: Currently only 'heroku' and 'clevercloud' is supported as a source")
		os.Exit(1)
	}

	// Validate destination
	if destination != "aws" && destination != "gcp" && destination != "scaleway" {
		fmt.Println("Error: Destination must be 'aws', 'gcp', or 'scaleway'")
		os.Exit(1)
	}

	// Check for HEROKU_API_KEY when source is heroku
	herokuAPIKey := os.Getenv("HEROKU_API_KEY")
	if source == "heroku" && herokuAPIKey == "" {
		fmt.Println("Error: HEROKU_API_KEY env var must be set when using Heroku as the source")
		os.Exit(1)
	}

	// Check for CLEVERCLOUD_AUTH_TOKEN when source is clevercloud
	clevercloudAuthToken := os.Getenv("CLEVERCLOUD_AUTH_TOKEN")
	if source == "clevercloud" && clevercloudAuthToken == "" {
		fmt.Println("Error: CLEVERCLOUD_AUTH_TOKEN env var must be set when using Clever Cloud as the source")
		os.Exit(1)
	}

	// Check for AWS credentials for Bedrock
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("AWS_REGION")

	if awsAccessKey == "" || awsSecretKey == "" {
		fmt.Println("Error: AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set in the environment")
		os.Exit(1)
	}

	if awsRegion == "" {
		awsRegion = "us-east-1" // Default to us-east-1 if not specified
		fmt.Println("Warning: AWS_REGION not set, defaulting to us-east-1")
	}

	// Check for Bedrock Model ARN
	bedrockModelARN := os.Getenv("AWS_BEDROCK_MODEL_ARN")
	if bedrockModelARN == "" {
		fmt.Println("Error: AWS_BEDROCK_MODEL_ARN must be set in the environment")
		os.Exit(1)
	}

	qoveryAPIKey := os.Getenv("QOVERY_API_KEY")
	if qoveryAPIKey == "" {
		fmt.Println("Error: QOVERY_API_KEY must be set in the environment")
		os.Exit(1)
	}

	githubToken := os.Getenv("GITHUB_TOKEN") // optional

	// Create a progress channel
	progressChan := make(chan migration.ProgressUpdate)

	// Create a new progress bar
	bar := progressbar.NewOptions(100,
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("Preparing..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "🚀",
			SaucerHead:    "🚀",
			SaucerPadding: "·",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// Start a goroutine to update the progress bar
	go func() {
		for update := range progressChan {
			_ = bar.Set(int(update.Progress * 100))
			bar.Describe(update.Stage)
		}
	}()

	bedrockClientConfig := bedrock.DefaultConfig()
	bedrockClientConfig.AWSRegion = awsRegion
	bedrockClientConfig.InferenceProfileARN = bedrockModelARN

	var assets *migration.Assets
	var err error

	if source == "heroku" {
		assets, err = migration.GenerateHerokuMigrationAssets(
			herokuAPIKey,
			awsAccessKey,
			awsSecretKey,
			qoveryAPIKey,
			githubToken,
			destination,
			bedrockClientConfig,
			progressChan,
		)
	}

	if source == "clevercloud" {
		assets, err = migration.GenerateCleverCloudMigrationAssets(
			clevercloudAuthToken,
			awsAccessKey,
			awsSecretKey,
			qoveryAPIKey,
			githubToken,
			destination,
			bedrockClientConfig,
			progressChan,
		)
	}

	// Close the progress channel
	close(progressChan)

	// Ensure the progress bar reaches 100%
	_ = bar.Finish()

	if err != nil {
		fmt.Printf("\nError generating migration assets: %v\n", err)
		os.Exit(1)
	}

	if outputDir != "" {
		err := migration.WriteAssets(outputDir, assets, true)
		if err != nil {
			fmt.Printf("\nError writing migration assets: %v\n", err)
			return
		}
		fmt.Printf("\nMigration assets prepared successfully in %s\n", outputDir)
		return
	}

	// Output the generated assets to stdout
	for _, generatedTfFile := range assets.GeneratedTerraformFiles {
		fmt.Println("====================================")
		fmt.Println(generatedTfFile.AppName)
		fmt.Println("====================================")

		fmt.Println("\nTerraform Main Configuration:")
		fmt.Println(generatedTfFile.MainTf)

		fmt.Println("\nTerraform Variables:")
		fmt.Println(generatedTfFile.VariablesTf)

		// output Dockerfile content if it exists
		for _, dockerfile := range assets.Dockerfiles {
			if dockerfile.AppName == generatedTfFile.AppName {
				fmt.Println("\nDockerfile:")
				fmt.Println(dockerfile.DockerfileContent)
				break
			}
		}
	}
}
