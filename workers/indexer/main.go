package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/data-storage-manager/workers/common"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

// NewsMessage - Formato del scraper (campos en español)
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
	db      *sql.DB
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *common.Config
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
	db, err := sql.Open("postgres", cfg.GetPostgresConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	_, err = ch.QueueDeclare("ingestion_queue", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare ingestion queue: %v", err)
	}

	_, err = ch.QueueDeclare("sync_queue", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare sync queue: %v", err)
	}

	log.Println("Connected to PostgreSQL and RabbitMQ")

	return &Indexer{
		db:      db,
		conn:    conn,
		channel: ch,
		config:  cfg,
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

		// Zona horaria de Chile (UTC-3/-4)
		loc, _ := time.LoadLocation("America/Santiago")
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

func (idx *Indexer) getOrCreateTag(tagName string) (string, error) {
	var tagID string

	err := idx.db.QueryRow(
		"SELECT id FROM tags WHERE name = $1",
		tagName,
	).Scan(&tagID)

	if err == sql.ErrNoRows {
		err = idx.db.QueryRow(
			"INSERT INTO tags (name) VALUES ($1) RETURNING id",
			tagName,
		).Scan(&tagID)
	}

	return tagID, err
}

func (idx *Indexer) saveNews(msg NewsMessage, urlHash, contentHash, mediaSourceID string, publishedDate time.Time) (string, error) {
	var newsID string

	err := idx.db.QueryRow(`
		INSERT INTO news (
			title, content, abstract, author, author_description,
			media_source_id, published_date, url, url_hash, content_hash, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, msg.Titulo, msg.Cuerpo, msg.Abstract, msg.Autor, msg.DescAutor,
		mediaSourceID, publishedDate, msg.URL, urlHash, contentHash, "indexed").Scan(&newsID)

	if err != nil {
		return "", err
	}

	// Guardar multimedia con tipo
	mediaType := msg.TipoMultimedia
	if mediaType == "" {
		mediaType = "imagen"
	}

	for _, mediaURL := range msg.Multimedia {
		_, err := idx.db.Exec(`
			INSERT INTO news_multimedia (news_id, url, media_type)
			VALUES ($1, $2, $3)
		`, newsID, mediaURL, mediaType)

		if err != nil {
			log.Printf("⚠ Warning: failed to insert multimedia: %v", err)
		}
	}

	// Guardar tags
	for _, tagName := range msg.Tags {
		tagID, err := idx.getOrCreateTag(tagName)
		if err != nil {
			log.Printf("⚠ Warning: failed to get/create tag '%s': %v", tagName, err)
			continue
		}

		_, err = idx.db.Exec(`
			INSERT INTO news_tags (news_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, newsID, tagID)

		if err != nil {
			log.Printf("⚠ Warning: failed to associate tag '%s': %v", tagName, err)
		}
	}

	return newsID, nil
}

func (idx *Indexer) processMessage(msg amqp.Delivery) error {
	var news NewsMessage

	if err := json.Unmarshal(msg.Body, &news); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false)
		return err
	}

	log.Printf("→ Processing job_id: %s | URL: %s", news.JobID, news.URL)

	// Normalizar URL
	normalizedURL, err := idx.normalizeURL(news.URL)
	if err != nil {
		log.Printf("Failed to normalize URL: %v", err)
		msg.Nack(false, false)
		return err
	}
	news.URL = normalizedURL

	// Parsear fecha española
	publishedDate, err := idx.parseSpanishDate(news.Fecha)
	if err != nil {
		log.Printf("Failed to parse date '%s': %v", news.Fecha, err)
		msg.Nack(false, false)
		return err
	}

	// Generar hashes
	urlHash := idx.generateHash(normalizedURL)
	contentHash := idx.generateHash(news.Titulo + news.Cuerpo)

	// Verificar duplicados
	duplicate, err := idx.isDuplicate(urlHash)
	if err != nil {
		log.Printf("Database error: %v", err)
		msg.Nack(false, true)
		return err
	}

	if duplicate {
		log.Printf("Duplicate detected, discarding message")
		msg.Ack(false)
		return nil
	}

	// Extraer media outlet de la URL
	mediaOutlet := idx.extractMediaOutlet(news.URL)
	country := "chile" // Default

	// Obtener o crear media source
	mediaSourceID, err := idx.getOrCreateMediaSource(mediaOutlet, country)
	if err != nil {
		log.Printf("Failed to get/create media source: %v", err)
		msg.Nack(false, true)
		return err
	}

	// Guardar noticia
	newsID, err := idx.saveNews(news, urlHash, contentHash, mediaSourceID, publishedDate)
	if err != nil {
		log.Printf("Failed to save news: %v", err)
		msg.Nack(false, true)
		return err
	}

	log.Printf("✓ News saved to PostgreSQL with ID: %s", newsID)

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
		log.Printf("Failed to publish to sync queue: %v", err)
		msg.Nack(false, true)
		return err
	}

	log.Printf("✓ Message sent to sync_queue")
	msg.Ack(false)
	return nil
}

func (idx *Indexer) Start() error {
	msgs, err := idx.channel.Consume("ingestion_queue", "indexer", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to consume: %v", err)
	}

	log.Println("Indexer worker started, waiting for messages...")

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			idx.processMessage(msg)
		}
	}()

	<-forever
	return nil
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
		log.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	if err := indexer.Start(); err != nil {
		log.Fatalf("Indexer error: %v", err)
	}
}
