package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/data-storage-manager/workers/common"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type NewsMessage struct {
	JobID          string   `json:"job_id"`
	URL            string   `json:"url"`
	Titulo         string   `json:"titulo"`
	Fecha          string   `json:"fecha"` // Spanish date format
	Tags           []string `json:"tags"`
	Autor          string   `json:"autor"`
	DescAutor      string   `json:"desc_autor"`
	Abstract       string   `json:"abstract"`
	Cuerpo         string   `json:"cuerpo"`
	Multimedia     []string `json:"multimedia"`
	TipoMultimedia string   `json:"tipo_multimedia"`
	ReceivedAt     time.Time `json:"received_at"`
}

type Indexer struct {
	db           *sql.DB
	conn         *amqp.Connection
	channel      *amqp.Channel
	config       *common.Config
	retryHandler *common.RetryHandler
	logger       *common.Logger
}

var (
	spanishMonths = map[string]int{
		"enero": 1, "febrero": 2, "marzo": 3, "abril": 4,
		"mayo": 5, "junio": 6, "julio": 7, "agosto": 8,
		"septiembre": 9, "octubre": 10, "noviembre": 11, "diciembre": 12,
	}
	spanishDatePattern = regexp.MustCompile(`(?i)(?:lunes|martes|miércoles|jueves|viernes|sábado|domingo)?\s*(\d{1,2})\s+(\w+)\s+de\s+(\d{4})\s*\|\s*(\d{1,2}):(\d{2})`)
)

func NewIndexer(cfg *common.Config) (*Indexer, error) {
	// Crear logger estructurado
	logger := common.NewLogger("worker-indexer")

	logger.Info("Initializing indexer worker")

	db, err := sql.Open("postgres", cfg.GetPostgresConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	logger.Info("Connected to PostgreSQL")

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	logger.Info("Connected to RabbitMQ")

	// Configurar retry handler con DLQ
	retryHandler, err := common.NewRetryHandler(ch, "ingestion_queue", 3)
	if err != nil {
		return nil, fmt.Errorf("failed to setup retry handler: %v", err)
	}

	// NOTA: La cola sync_queue es declarada por el worker-sync con DLX configurado
	// El worker-indexer solo publica mensajes, no necesita declarar la cola

	logger.Info("Indexer worker initialized successfully")

	return &Indexer{
		db:           db,
		conn:         conn,
		channel:      ch,
		config:       cfg,
		retryHandler: retryHandler,
		logger:       logger,
	}, nil
}

// parseSpanishDate convierte fecha española a time.Time
// Ejemplo: "Martes 16 septiembre de 2025 | 23:01" -> time.Time
func (idx *Indexer) parseSpanishDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)

	matches := spanishDatePattern.FindStringSubmatch(dateStr)
	if len(matches) == 6 {
		day, _ := strconv.Atoi(matches[1])
		monthStr := strings.ToLower(matches[2])
		year, _ := strconv.Atoi(matches[3])
		hour, _ := strconv.Atoi(matches[4])
		minute, _ := strconv.Atoi(matches[5])

		month, exists := spanishMonths[monthStr]
		if !exists {
			return time.Time{}, fmt.Errorf("mes no reconocido: %s", monthStr)
		}

		loc, err := time.LoadLocation("America/Santiago")
		if err != nil {
			loc = time.UTC
		}
		date := time.Date(year, time.Month(month), day, hour, minute, 0, 0, loc)
		return date, nil
	}

	return time.Time{}, fmt.Errorf("formato de fecha no reconocido: %s", dateStr)
}

// extractMediaOutlet extrae el medio desde la URL
func (idx *Indexer) extractMediaOutlet(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "unknown"
	}

	domain := strings.TrimPrefix(parsed.Host, "www.")
	parts := strings.Split(domain, ".")
	if len(parts) > 0 {
		return parts[0]
	}

	return "unknown"
}

func (idx *Indexer) normalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	normalized := strings.ToLower(parsedURL.Scheme + "://" + strings.TrimPrefix(parsedURL.Host, "www.") + parsedURL.Path)
	normalized = strings.TrimSuffix(normalized, "/")
	return normalized, nil
}

func (idx *Indexer) generateHash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (idx *Indexer) isDuplicate(urlHash string) (bool, error) {
	var count int
	err := idx.db.QueryRow("SELECT COUNT(*) FROM news WHERE url_hash = $1", urlHash).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (idx *Indexer) getOrCreateMediaSource(name, country string) (string, error) {
	var sourceID string

	err := idx.db.QueryRow(
		"SELECT id FROM media_sources WHERE name = $1",
		name,
	).Scan(&sourceID)

	if err == sql.ErrNoRows {
		err = idx.db.QueryRow(
			"INSERT INTO media_sources (name, country) VALUES ($1, $2) RETURNING id",
			name, country,
		).Scan(&sourceID)
	}

	return sourceID, err
}

func (idx *Indexer) saveNews(msg NewsMessage, urlHash, contentHash, mediaSourceID string, publishedDate time.Time) (string, error) {
	var newsID string

	// Iniciar transacción
	tx, err := idx.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Defer para asegurar rollback en caso de pánico o error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic después de rollback
		}
	}()

	// 1. INSERT news
	err = tx.QueryRow(`
		INSERT INTO news (
			title, content, abstract, author, author_description,
			media_source_id, published_date, url, url_hash, content_hash, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, msg.Titulo, msg.Cuerpo, msg.Abstract, msg.Autor, msg.DescAutor,
		mediaSourceID, publishedDate, msg.URL, urlHash, contentHash, "indexed").Scan(&newsID)

	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to insert news: %v", err)
	}

	idx.logger.WithNewsID(newsID).Debug("News record created in transaction")

	// 2. INSERT multimedia con tipo
	mediaType := msg.TipoMultimedia
	if mediaType == "" {
		mediaType = "imagen"
	}

	for i, mediaURL := range msg.Multimedia {
		_, err := tx.Exec(`
			INSERT INTO news_multimedia (news_id, url, media_type)
			VALUES ($1, $2, $3)
		`, newsID, mediaURL, mediaType)

		if err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to insert multimedia #%d (%s): %v", i+1, mediaURL, err)
		}
	}

	if len(msg.Multimedia) > 0 {
		idx.logger.WithFields(logrus.Fields{
			"news_id":          newsID,
			"multimedia_count": len(msg.Multimedia),
		}).Debug("Multimedia items inserted")
	}

	// 3. INSERT tags (dentro de la transacción)
	tagIDs := []string{}
	for _, tagName := range msg.Tags {
		// Primero intentar obtener el tag (usando la transacción)
		var tagID string
		err := tx.QueryRow("SELECT id FROM tags WHERE name = $1", tagName).Scan(&tagID)

		if err == sql.ErrNoRows {
			// Tag no existe, crearlo dentro de la transacción
			err = tx.QueryRow(
				"INSERT INTO tags (name) VALUES ($1) RETURNING id",
				tagName,
			).Scan(&tagID)

			if err != nil {
				tx.Rollback()
				return "", fmt.Errorf("failed to create tag '%s': %v", tagName, err)
			}
		} else if err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to query tag '%s': %v", tagName, err)
		}

		tagIDs = append(tagIDs, tagID)
	}

	// 4. INSERT news_tags (relaciones)
	for i, tagID := range tagIDs {
		_, err = tx.Exec(`
			INSERT INTO news_tags (news_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, newsID, tagID)

		if err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to associate tag #%d: %v", i+1, err)
		}
	}

	if len(tagIDs) > 0 {
		idx.logger.WithFields(logrus.Fields{
			"news_id":    newsID,
			"tags_count": len(tagIDs),
		}).Debug("Tags associated")
	}

	// COMMIT: Si llegamos aquí, todo fue exitoso
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %v", err)
	}

	idx.logger.WithNewsID(newsID).Info("Transaction committed successfully")
	return newsID, nil
}

func (idx *Indexer) processMessage(msg amqp.Delivery) error {
	startTime := time.Now()
	var news NewsMessage

	// Registrar mensaje recibido
	idx.logger.RecordMessageProcessed()

	// Intentar unmarshal del JSON
	if err := json.Unmarshal(msg.Body, &news); err != nil {
		idx.logger.WithError(err).Error("Failed to unmarshal message")
		idx.logger.RecordMessageFailure("unmarshal_error")
		// Error no recuperable (mensaje malformado) - enviar a DLQ
		idx.retryHandler.HandleError(msg, err, false)
		return err
	}

	logger := idx.logger.WithJobID(news.JobID)
	logger.WithFields(logrus.Fields{
		"url":   news.URL,
		"title": news.Titulo,
	}).Info("Processing message")

	// Normalizar URL
	normalizedURL, err := idx.normalizeURL(news.URL)
	if err != nil {
		logger.WithError(err).Error("Failed to normalize URL")
		idx.logger.RecordMessageFailure("url_normalization_error")
		// Error no recuperable (URL inválida) - enviar a DLQ
		idx.retryHandler.HandleError(msg, err, false)
		return err
	}
	news.URL = normalizedURL

	// Parsear fecha española
	publishedDate, err := idx.parseSpanishDate(news.Fecha)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"fecha": news.Fecha,
			"error": err.Error(),
		}).Error("Failed to parse date")
		idx.logger.RecordMessageFailure("date_parse_error")
		// Error no recuperable (formato de fecha inválido) - enviar a DLQ
		idx.retryHandler.HandleError(msg, err, false)
		return err
	}

	// Generar hashes
	urlHash := idx.generateHash(normalizedURL)
	contentHash := idx.generateHash(news.Titulo + news.Cuerpo)

	// Verificar duplicados
	duplicate, err := idx.isDuplicate(urlHash)
	if err != nil {
		logger.WithError(err).Error("Database error checking duplicates")
		idx.logger.RecordMessageFailure("db_duplicate_check_error")
		// Error recuperable (problema de DB) - reintentar
		idx.logger.RecordMessageRetry()
		idx.retryHandler.HandleError(msg, err, true)
		return err
	}

	if duplicate {
		logger.Warn("Duplicate detected, discarding message")
		idx.retryHandler.AckSuccess(msg)
		// No contar como éxito ni fallo, es un caso esperado
		return nil
	}

	// Extraer media outlet de la URL
	mediaOutlet := idx.extractMediaOutlet(news.URL)
	country := "chile" // Default

	// Obtener o crear media source
	mediaSourceID, err := idx.getOrCreateMediaSource(mediaOutlet, country)
	if err != nil {
		logger.WithError(err).Error("Failed to get/create media source")
		idx.logger.RecordMessageFailure("media_source_error")
		// Error recuperable (problema de DB) - reintentar
		idx.logger.RecordMessageRetry()
		idx.retryHandler.HandleError(msg, err, true)
		return err
	}

	// Guardar noticia
	newsID, err := idx.saveNews(news, urlHash, contentHash, mediaSourceID, publishedDate)
	if err != nil {
		logger.WithError(err).Error("Failed to save news")
		idx.logger.RecordMessageFailure("save_news_error")
		// Error recuperable (problema de DB) - reintentar
		idx.logger.RecordMessageRetry()
		idx.retryHandler.HandleError(msg, err, true)
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"news_id":      newsID,
		"media_outlet": mediaOutlet,
	})
	logger.Info("News saved to PostgreSQL")

	// Publicar a sync_queue para Elasticsearch
	syncMessage := map[string]interface{}{
		"news_id": newsID,
		"action":  "index",
	}

	body, _ := json.Marshal(syncMessage)
	err = idx.channel.Publish("", "sync_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	if err != nil {
		logger.WithError(err).Error("Failed to publish to sync queue")
		idx.logger.RecordMessageFailure("sync_queue_publish_error")
		// Error recuperable (problema con RabbitMQ) - reintentar
		idx.logger.RecordMessageRetry()
		idx.retryHandler.HandleError(msg, err, true)
		return err
	}

	duration := time.Since(startTime)
	logger.WithFields(logrus.Fields{
		"duration_ms": duration.Milliseconds(),
	}).Info("Message processed successfully")

	// Registrar métricas de éxito
	idx.logger.RecordMessageSuccess(duration)

	idx.retryHandler.AckSuccess(msg)
	return nil
}

func (idx *Indexer) Start() error {
	msgs, err := idx.channel.Consume("ingestion_queue", "indexer", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to consume: %v", err)
	}

	idx.logger.Info("Indexer worker started, waiting for messages...")

	forever := make(chan bool)

	// Goroutine para procesar mensajes
	go func() {
		for msg := range msgs {
			idx.processMessage(msg)
		}
	}()

	// Goroutine para loguear métricas cada 5 minutos
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			idx.logger.LogMetrics()
		}
	}()

	// Iniciar servidor HTTP para métricas
	idx.startMetricsServer()

	<-forever
	return nil
}

func (idx *Indexer) startMetricsServer() {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := idx.logger.GetMetrics()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "worker-indexer",
		})
	})

	go func() {
		idx.logger.Info("Metrics server starting on :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			idx.logger.WithError(err).Error("Metrics server failed")
		}
	}()
}

func (idx *Indexer) Close() {
	if idx.channel != nil {
		idx.channel.Close()
	}
	if idx.conn != nil {
		idx.conn.Close()
	}
	if idx.db != nil {
		idx.db.Close()
	}
}

func main() {
	cfg := common.LoadConfig()

	indexer, err := NewIndexer(cfg)
	if err != nil {
		// Usar logger básico si falla la inicialización
		logrus.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	if err := indexer.Start(); err != nil {
		indexer.logger.WithError(err).Fatal("Indexer error")
	}
}
