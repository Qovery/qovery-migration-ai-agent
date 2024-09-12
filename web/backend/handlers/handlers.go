package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type MigrationRequest struct {
	Source      string            `json:"source"`
	Destination string            `json:"destination"`
	Credentials map[string]string `json:"credentials"`
}

func MigrateHandler(c *gin.Context) {
	var req MigrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use your Go library to generate Terraform manifests and Dockerfiles
	terraformManifest, dockerfile, err := qovery_migration_library.GenerateMigrationFiles(req.Source, req.Destination, req.Credentials)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate migration files"})
		return
	}

	// Save generated files (in a real-world scenario, consider using a more secure method)
	// This is a simplified example and may not be suitable for production use
	if err := saveFile("terraform.tf", terraformManifest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save Terraform manifest"})
		return
	}
	if err := saveFile("Dockerfile", dockerfile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save Dockerfile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Migration files generated successfully"})
}

func DownloadHandler(c *gin.Context) {
	filename := c.Param("filename")
	if filename != "terraform.tf" && filename != "Dockerfile" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}

	filepath := filepath.Join(".", filename)
	c.File(filepath)
}

func saveFile(filename string, content string) error {
	// Implement file saving logic
	// This is a placeholder and should be replaced with actual file writing code
	return nil
}
