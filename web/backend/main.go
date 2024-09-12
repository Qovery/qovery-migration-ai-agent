package main

import (
	"backend/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "https://migrate.qovery.com"}
	r.Use(cors.New(config))

	// Routes
	r.POST("/api/migrate", handlers.MigrateHandler)
	r.GET("/api/download/:filename", handlers.DownloadHandler)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
