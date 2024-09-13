package cmd

import (
	"fmt"
	"github.com/qovery/qovery-migration-ai-agent"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	source      string
	destination string
	outputDir   string
	stdoutFlag  bool
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
	prepareCmd.Flags().StringVarP(&source, "from", "f", "", "Source platform (e.g., 'heroku') (required)")
	prepareCmd.Flags().StringVarP(&destination, "to", "t", "", "Destination cloud provider (aws, gcp, or scaleway) (required)")
	prepareCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for generated files")
	_ = prepareCmd.MarkFlagRequired("from")
	_ = prepareCmd.MarkFlagRequired("to")
}

func runPrepare(cmd *cobra.Command, args []string) {
	// Validate source
	if source != "heroku" {
		fmt.Println("Error: Currently only 'heroku' is supported as a source")
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
		fmt.Println("Error: HEROKU_API_KEY must be set in the .env file when using Heroku as the source")
		os.Exit(1)
	}

	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")
	qoveryAPIKey := os.Getenv("QOVERY_API_KEY")

	if claudeAPIKey == "" || qoveryAPIKey == "" {
		fmt.Println("Error: CLAUDE_API_KEY and QOVERY_API_KEY must be set in the .env file")
		os.Exit(1)
	}

	// Create a progress channel
	progressChan := make(chan qovery_migration_ai_agent.migration.ProgressUpdate)

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

	assets, err := migration.GenerateMigrationAssets(source, herokuAPIKey, claudeAPIKey, qoveryAPIKey, destination, progressChan)

	// Close the progress channel
	close(progressChan)

	// Ensure the progress bar reaches 100%
	_ = bar.Finish()

	if err != nil {
		fmt.Printf("\nError generating migration assets: %v\n", err)
		os.Exit(1)
	}

	if outputDir != "" {
		// Create the output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		// Write Terraform main configuration
		if err := writeToFile(filepath.Join(outputDir, "main.tf"), assets.TerraformMain); err != nil {
			fmt.Printf("Error writing main.tf: %v\n", err)
			os.Exit(1)
		}

		// Write Terraform variables
		if err := writeToFile(filepath.Join(outputDir, "variables.tf"), assets.TerraformVariables); err != nil {
			fmt.Printf("Error writing variables.tf: %v\n", err)
			os.Exit(1)
		}

		// Write Dockerfiles
		for _, dockerfile := range assets.Dockerfiles {
			filename := fmt.Sprintf("Dockerfile-%s", dockerfile.AppName)
			if err := writeToFile(filepath.Join(outputDir, filename), dockerfile.DockerfileContent); err != nil {
				fmt.Printf("Error writing %s: %v\n", filename, err)
				os.Exit(1)
			}
		}

		fmt.Printf("\nMigration assets prepared successfully in %s\n", outputDir)
		return
	}

	// Output the generated assets to stdout
	fmt.Println("\nTerraform Main Configuration:")
	fmt.Println(assets.TerraformMain)
	fmt.Println("\nTerraform Variables:")
	fmt.Println(assets.TerraformVariables)
	fmt.Println("\nDockerfiles:")
	for _, dockerfile := range assets.Dockerfiles {
		fmt.Printf("App: %s\n", dockerfile.AppName)
		fmt.Println(dockerfile.DockerfileContent)
		fmt.Println("---")
	}
}

func writeToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
