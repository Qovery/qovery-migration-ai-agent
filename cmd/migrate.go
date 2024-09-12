package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"qovery-ai-migration/pkg/migration"

	"github.com/spf13/cobra"
)

var (
	source      string
	destination string
	outputDir   string
	stdoutFlag  bool
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate applications to Qovery",
	Long:  `This command migrates your applications to Qovery, generating necessary Terraform configurations and Dockerfiles.`,
	Run:   runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&source, "from", "f", "", "Source platform (e.g., 'heroku') (required)")
	migrateCmd.Flags().StringVarP(&destination, "to", "t", "", "Destination cloud provider (aws, gcp, or scaleway) (required)")
	migrateCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for generated files")
	_ = migrateCmd.MarkFlagRequired("from")
	_ = migrateCmd.MarkFlagRequired("to")
}

func runMigrate(cmd *cobra.Command, args []string) {
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

	assets, err := migration.GenerateMigrationAssets(source, herokuAPIKey, claudeAPIKey, qoveryAPIKey, destination)
	if err != nil {
		fmt.Printf("Error generating migration assets: %v\n", err)
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

		fmt.Printf("Migration assets generated successfully in %s\n", outputDir)
		return
	}

	// Output the generated assets to stdout
	fmt.Println("Terraform Main Configuration:")
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
