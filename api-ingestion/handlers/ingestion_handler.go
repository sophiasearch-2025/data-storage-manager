package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/newspress/api-ingestion/models"
	"github.com/newspress/api-ingestion/services"
	"github.com/newspress/api-ingestion/utils"
)

type IngestionHandler struct {
	rabbitMQ *services.RabbitMQService
}

func NewIngestionHandler(rabbitMQ *services.RabbitMQService) *IngestionHandler {
	return &IngestionHandler{rabbitMQ: rabbitMQ}
}

func (h *IngestionHandler) IngestNews(c *gin.Context) {
	var req models.NewsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	jobID := uuid.New().String()

	// El mensaje se envía con el formato original del scraper
	message := map[string]interface{}{
		"job_id":          jobID,
		"url":             req.URL,
		"titulo":          req.Titulo,
		"fecha":           req.Fecha, // En formato español
		"tags":            req.Tags,
		"autor":           req.Autor,
		"desc_autor":      req.DescAutor,
		"abstract":        req.Abstract,
		"cuerpo":          req.Cuerpo,
		"multimedia":      req.Multimedia,
		"tipo_multimedia": req.TipoMultimedia,
		"received_at":     time.Now(),
	}

	if err := h.rabbitMQ.PublishNews(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "queue_error",
			"message": "failed to publish message to queue",
		})
		return
	}

	c.JSON(http.StatusAccepted, models.NewsResponse{
		JobID:   jobID,
		Status:  "pending",
		Message: "news queued for processing",
	})
}

func (h *IngestionHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "api-ingestion",
	})
}
