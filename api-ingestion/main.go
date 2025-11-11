package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/data-storage-manager/api-ingestion/config"
	"github.com/data-storage-manager/api-ingestion/handlers"
	"github.com/data-storage-manager/api-ingestion/services"
)

func main() {
	cfg := config.Load()

	rabbitMQ, err := services.NewRabbitMQService(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	ingestionHandler := handlers.NewIngestionHandler(rabbitMQ)

	router := gin.Default()

	router.GET("/health", ingestionHandler.Health)
	router.POST("/api/v1/news", ingestionHandler.IngestNews)

	log.Printf("API Ingestion listening on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
