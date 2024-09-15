package main

import (
	"backend/handlers"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func main() {
	config := handlers.Config{
		ClaudeAPIKey: os.Getenv("CLAUDE_API_KEY"),
		QoveryAPIKey: os.Getenv("QOVERY_API_KEY"),
		GitHubToken:  os.Getenv("GITHUB_TOKEN"),
	}

	r := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{
		"*", // Allow all origins - temp
		//"http://localhost:3000",
		//"https://migrate.qovery.com",
		//os.Getenv("FRONTEND_HOST_URL"),
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// Routes
	r.POST("/api/migrate/heroku", handlers.HerokuMigrateHandler(config))

	// Handle preflight requests
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Get port from environment variable or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // Default port if SERVER_PORT is not set
	}

	// Start server
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
