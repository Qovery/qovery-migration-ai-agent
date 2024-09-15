package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Qovery/qovery-migration-ai-agent/pkg/migration"
	"github.com/gin-gonic/gin"
)

type Config struct {
	ClaudeAPIKey string
	QoveryAPIKey string
	GitHubToken  string
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
		defer os.RemoveAll(tempDir) // Clean up the temporary directory when we're done

		progressChan := make(chan migration.ProgressUpdate)
		go func() {
			for update := range progressChan {
				// You can use this to send real-time updates to the client
				// For example, you could use WebSockets or Server-Sent Events
				_ = update // Placeholder to avoid unused variable error
			}
		}()

		// Use your Go library to generate Terraform manifests and Dockerfiles
		assets, err := migration.GenerateMigrationAssets(req.Source, req.HerokuAPIKey, config.ClaudeAPIKey,
			config.QoveryAPIKey, config.GitHubToken, req.Destination, progressChan)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Write the generated assets to the temporary directory
		err = migration.WriteAssets(tempDir, assets)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create a zip file
		zipPath := filepath.Join(tempDir, "heroku-migration.zip")
		err = createZip(tempDir, zipPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create zip file"})
			return
		}

		// Set the appropriate headers for file download
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename=heroku-migration.zip")
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
