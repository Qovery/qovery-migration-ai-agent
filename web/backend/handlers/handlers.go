package handlers

import (
	"archive/zip"
	"backend/services"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Qovery/qovery-migration-ai-agent/pkg/bedrock"
	"github.com/Qovery/qovery-migration-ai-agent/pkg/migration"
	"github.com/gin-gonic/gin"
)

type Config struct {
	QoveryAPIKey           string
	GitHubToken            string
	S3AccessKeyId          string
	S3SecretAccessKey      string
	S3Bucket               string
	S3Region               string
	BedrockAccessKeyId     string
	BedrockSecretAccessKey string
	BedrockRegion          string
}

type HerokuMigrationRequest struct {
	Source       string `json:"source"`
	Destination  string `json:"destination"`
	HerokuAPIKey string `json:"herokuApiKey"`
}

func HerokuMigrateHandler(config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req HerokuMigrationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.HerokuAPIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Heroku API Key is required"})
			return
		}

		// Create a temporary directory
		tempDir, err := ioutil.TempDir("", "heroku-migration-")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary directory"})
			return
		}
		defer os.RemoveAll(tempDir)

		progressChan := make(chan migration.ProgressUpdate)
		go func() {
			for update := range progressChan {
				// You can use this to send real-time updates to the client
				_ = update // Placeholder to avoid unused variable error
			}
		}()

		// Create Bedrock client configuration
		bedrockClientConfig := bedrock.DefaultConfig()
		bedrockClientConfig.AWSRegion = config.BedrockRegion

		// Use your Go library to generate Terraform manifests and Dockerfiles
		assets, err := migration.GenerateHerokuMigrationAssets(
			req.HerokuAPIKey,
			config.BedrockAccessKeyId,
			config.BedrockSecretAccessKey,
			config.QoveryAPIKey,
			config.GitHubToken,
			req.Destination,
			bedrockClientConfig,
			progressChan,
		)

		if err != nil {
			// Error occurred, let's zip the assets and upload to S3
			errorZipName := fmt.Sprintf("error-heroku-migration-%s.zip", time.Now().Format("20060102-150405"))
			errorZipPath := filepath.Join(tempDir, errorZipName)

			// Write the assets to the temporary directory
			if writeErr := migration.WriteAssets(tempDir, assets, true); writeErr == nil {
				// Create a zip file
				if zipErr := createZip(tempDir, errorZipPath); zipErr == nil {
					// Upload the error zip file to S3
					if config.S3Bucket != "" && config.S3Region != "" {
						_, uploadErr := services.UploadZipToS3(errorZipPath, config.S3Bucket, config.S3Region, config.S3AccessKeyId, config.S3SecretAccessKey)
						if uploadErr != nil {
							fmt.Printf("Failed to upload error zip to S3: %v\n", uploadErr)
						} else {
							fmt.Printf("Error zip uploaded to S3: %s\n", errorZipPath)
						}
					}
				}
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Write the generated assets to the temporary directory
		err = migration.WriteAssets(tempDir, assets, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create a zip file
		zipName := fmt.Sprintf("heroku-migration-%s.zip", time.Now().Format("20060102-150405"))
		zipPath := filepath.Join(tempDir, zipName)
		err = createZip(tempDir, zipPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create zip file"})
			return
		}

		if config.S3Bucket != "" && config.S3Region != "" {
			// Upload the zip file to S3
			_, err := services.UploadZipToS3(zipPath, config.S3Bucket, config.S3Region, config.S3AccessKeyId, config.S3SecretAccessKey)
			if err != nil {
				fmt.Printf("Failed to upload zip to S3: %v\n", err)
			} else {
				fmt.Printf("Zip uploaded to S3: %s\n", zipPath)
			}
		}

		// Set the appropriate headers for file download
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipName))
		c.Header("Content-Type", "application/zip")

		// Send the file
		c.File(zipPath)
	}
}

func createZip(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the zip file itself
		if path == zipPath {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create file header: %v", err)
		}

		header.Name, err = filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create header: %v", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
