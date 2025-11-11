package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/data-storage-manager/workers/common"
	"github.com/elastic/go-elasticsearch/v8"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

type SyncMessage struct {
	NewsID string `json:"news_id"`
	Action string `json:"action"` // index, update, delete
}

type NewsDocument struct {
	NewsID            string      `json:"news_id"`
	Title             string      `json:"title"`
	Content           string      `json:"content"`
	Abstract          string      `json:"abstract"`
	Author            string      `json:"author"`
	AuthorDescription string      `json:"author_description"`
	MediaSource       MediaSource `json:"media_source"`
	PublishedDate     time.Time   `json:"published_date"`
	URL               string      `json:"url"`
	Multimedia        []string    `json:"multimedia"`
	Tags              []string    `json:"tags"`
	Status            string      `json:"status"`
	IndexedAt         time.Time   `json:"indexed_at"`
}

type MediaSource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

type SyncWorker struct {
	db           *sql.DB
	es           *elasticsearch.Client
	conn         *amqp.Connection
	channel      *amqp.Channel
	config       *common.Config
	retryHandler *common.RetryHandler
}

func NewSyncWorker(cfg *common.Config) (*SyncWorker, error) {
	// Connect to PostgreSQL
	db, err := sql.Open("postgres", cfg.GetPostgresConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	// Connect to Elasticsearch
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchURL},
	}
	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %v", err)
	}

	// Test Elasticsearch connection
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get Elasticsearch info: %v", err)
	}
	defer res.Body.Close()

	// Connect to RabbitMQ
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// Configurar retry handler con DLQ para sync_queue
	retryHandler, err := common.NewRetryHandler(ch, "sync_queue", 3)
	if err != nil {
		return nil, fmt.Errorf("failed to setup retry handler: %v", err)
	}

	log.Println("Connected to PostgreSQL, Elasticsearch and RabbitMQ")

	return &SyncWorker{
		db:           db,
		es:           es,
		conn:         conn,
		channel:      ch,
		config:       cfg,
		retryHandler: retryHandler,
	}, nil
}

func (sw *SyncWorker) fetchNewsFromDB(newsID string) (*NewsDocument, error) {
	query := `
		SELECT
			n.id, n.title, n.content, n.abstract, n.author, n.author_description,
			n.published_date, n.url, n.status,
			ms.id, ms.name, ms.country
		FROM news n
		LEFT JOIN media_sources ms ON n.media_source_id = ms.id
		WHERE n.id = $1
	`

	var doc NewsDocument
	var ms MediaSource

	err := sw.db.QueryRow(query, newsID).Scan(
		&doc.NewsID, &doc.Title, &doc.Content, &doc.Abstract, &doc.Author, &doc.AuthorDescription,
		&doc.PublishedDate, &doc.URL, &doc.Status,
		&ms.ID, &ms.Name, &ms.Country,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %v", err)
	}

	doc.MediaSource = ms

	// Fetch multimedia
	multimediaQuery := `SELECT url FROM news_multimedia WHERE news_id = $1`
	rows, err := sw.db.Query(multimediaQuery, newsID)
	if err != nil {
		log.Printf("Warning: failed to fetch multimedia: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var url string
			if err := rows.Scan(&url); err == nil {
				doc.Multimedia = append(doc.Multimedia, url)
			}
		}
	}

	// Fetch tags
	tagsQuery := `
		SELECT t.name
		FROM tags t
		INNER JOIN news_tags nt ON t.id = nt.tag_id
		WHERE nt.news_id = $1
	`
	rows, err = sw.db.Query(tagsQuery, newsID)
	if err != nil {
		log.Printf("Warning: failed to fetch tags: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var tag string
			if err := rows.Scan(&tag); err == nil {
				doc.Tags = append(doc.Tags, tag)
			}
		}
	}

	doc.IndexedAt = time.Now()

	return &doc, nil
}

func (sw *SyncWorker) indexDocument(newsID string) error {
	doc, err := sw.fetchNewsFromDB(newsID)
	if err != nil {
		return err
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %v", err)
	}

	// Index document (using newsID as document ID for idempotency)
	res, err := sw.es.Index(
		"news",
		strings.NewReader(string(body)),
		sw.es.Index.WithDocumentID(newsID),
		sw.es.Index.WithRefresh("true"),
	)

	if err != nil {
		return fmt.Errorf("failed to index document: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned error: %s", res.String())
	}

	log.Printf("✓ Document indexed in Elasticsearch: %s", newsID)
	return nil
}

func (sw *SyncWorker) deleteDocument(newsID string) error {
	res, err := sw.es.Delete("news", newsID, sw.es.Delete.WithRefresh("true"))
	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("elasticsearch returned error: %s", res.String())
	}

	log.Printf("✓ Document deleted from Elasticsearch: %s", newsID)
	return nil
}

func (sw *SyncWorker) processMessage(msg amqp.Delivery) error {
	var syncMsg SyncMessage

	// Intentar unmarshal del JSON
	if err := json.Unmarshal(msg.Body, &syncMsg); err != nil {
		log.Printf("Failed to unmarshal sync message: %v", err)
		// Error no recuperable (mensaje malformado) - enviar a DLQ
		sw.retryHandler.HandleError(msg, err, false)
		return err
	}

	log.Printf("→ Processing sync: news_id=%s, action=%s", syncMsg.NewsID, syncMsg.Action)

	var err error
	switch syncMsg.Action {
	case "index", "update":
		err = sw.indexDocument(syncMsg.NewsID)
	case "delete":
		err = sw.deleteDocument(syncMsg.NewsID)
	default:
		log.Printf("Unknown action: %s", syncMsg.Action)
		// Error no recuperable (acción desconocida) - enviar a DLQ
		err = fmt.Errorf("unknown action: %s", syncMsg.Action)
		sw.retryHandler.HandleError(msg, err, false)
		return err
	}

	if err != nil {
		log.Printf("Sync failed: %v", err)
		// Error recuperable (problema con DB o ES) - reintentar
		sw.retryHandler.HandleError(msg, err, true)
		return err
	}

	sw.retryHandler.AckSuccess(msg)
	return nil
}

func (sw *SyncWorker) Start() error {
	msgs, err := sw.channel.Consume("sync_queue", "sync-worker", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to consume: %v", err)
	}

	log.Println("Sync worker started, waiting for messages...")

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			sw.processMessage(msg)
		}
	}()

	<-forever
	return nil
}

func (sw *SyncWorker) Close() {
	if sw.channel != nil {
		sw.channel.Close()
	}
	if sw.conn != nil {
		sw.conn.Close()
	}
	if sw.db != nil {
		sw.db.Close()
	}
}

func main() {
	cfg := common.LoadConfig()

	worker, err := NewSyncWorker(cfg)
	if err != nil {
		log.Fatalf("Failed to create sync worker: %v", err)
	}
	defer worker.Close()

	if err := worker.Start(); err != nil {
		log.Fatalf("Sync worker error: %v", err)
	}
}
