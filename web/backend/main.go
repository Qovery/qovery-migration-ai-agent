package main

import (
	"backend/handlers"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	config := handlers.Config{
		HerokuAPIKey:   os.Getenv("HEROKU_API_KEY"),
		ClaudeAPIKey:   os.Getenv("CLAUDE_API_KEY"),
		QoveryAPIKey:   os.Getenv("QOVERY_API_KEY"),
		AllowedOrigins: []string{"http://localhost:3000", "https://migrate.qovery.com"},
	}

	r := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = config.AllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	r.Use(cors.New(corsConfig))

	// Routes
	r.POST("/api/migrate", handlers.MigrateHandlerWithConfig(config))

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
