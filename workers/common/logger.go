package common

import (
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger es un wrapper sobre logrus con contexto
type Logger struct {
	*logrus.Logger
	serviceName string
	metrics     *Metrics
}

// Metrics almacena métricas básicas en memoria
type Metrics struct {
	mu                    sync.RWMutex
	messagesProcessed     int64
	messagesSucceeded     int64
	messagesFailed        int64
	messagesRetried       int64
	messagesDeadLettered  int64
	processingTimeTotal   time.Duration
	processingTimeCount   int64
	errorsByType          map[string]int64
	startTime             time.Time
}

// NewLogger crea un logger estructurado
func NewLogger(serviceName string) *Logger {
	log := logrus.New()

	// Configurar formato JSON para producción
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Configurar nivel de log desde variable de entorno
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		log.SetLevel(logrus.DebugLevel)
	case "WARN":
		log.SetLevel(logrus.WarnLevel)
	case "ERROR":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	log.SetOutput(os.Stdout)

	return &Logger{
		Logger:      log,
		serviceName: serviceName,
		metrics:     NewMetrics(),
	}
}

// WithFields añade campos estructurados al log
func (l *Logger) WithJobID(jobID string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"service": l.serviceName,
		"job_id":  jobID,
	})
}

// WithNewsID añade news_id al contexto
func (l *Logger) WithNewsID(newsID string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"service": l.serviceName,
		"news_id": newsID,
	})
}

// WithError añade error al contexto
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"service": l.serviceName,
		"error":   err.Error(),
	})
}

// WithContext añade múltiples campos
func (l *Logger) WithContext(fields map[string]interface{}) *logrus.Entry {
	fields["service"] = l.serviceName
	return l.WithFields(fields)
}

// === Métodos de Métricas ===

func NewMetrics() *Metrics {
	return &Metrics{
		errorsByType: make(map[string]int64),
		startTime:    time.Now(),
	}
}

// RecordMessageProcessed incrementa contador de mensajes procesados
func (l *Logger) RecordMessageProcessed() {
	l.metrics.mu.Lock()
	defer l.metrics.mu.Unlock()
	l.metrics.messagesProcessed++
}

// RecordMessageSuccess registra mensaje exitoso
func (l *Logger) RecordMessageSuccess(duration time.Duration) {
	l.metrics.mu.Lock()
	defer l.metrics.mu.Unlock()
	l.metrics.messagesSucceeded++
	l.metrics.processingTimeTotal += duration
	l.metrics.processingTimeCount++
}

// RecordMessageFailure registra mensaje fallido
func (l *Logger) RecordMessageFailure(errorType string) {
	l.metrics.mu.Lock()
	defer l.metrics.mu.Unlock()
	l.metrics.messagesFailed++
	l.metrics.errorsByType[errorType]++
}

// RecordMessageRetry registra reintento
func (l *Logger) RecordMessageRetry() {
	l.metrics.mu.Lock()
	defer l.metrics.mu.Unlock()
	l.metrics.messagesRetried++
}

// RecordMessageDeadLettered registra envío a DLQ
func (l *Logger) RecordMessageDeadLettered() {
	l.metrics.mu.Lock()
	defer l.metrics.mu.Unlock()
	l.metrics.messagesDeadLettered++
}

// GetMetrics retorna snapshot de métricas
func (l *Logger) GetMetrics() MetricsSnapshot {
	l.metrics.mu.RLock()
	defer l.metrics.mu.RUnlock()

	var avgProcessingTime time.Duration
	if l.metrics.processingTimeCount > 0 {
		avgProcessingTime = l.metrics.processingTimeTotal / time.Duration(l.metrics.processingTimeCount)
	}

	uptime := time.Since(l.metrics.startTime)

	// Copiar errorsByType para evitar race conditions
	errorsByType := make(map[string]int64)
	for k, v := range l.metrics.errorsByType {
		errorsByType[k] = v
	}

	return MetricsSnapshot{
		Service:              l.serviceName,
		Uptime:               uptime.String(),
		MessagesProcessed:    l.metrics.messagesProcessed,
		MessagesSucceeded:    l.metrics.messagesSucceeded,
		MessagesFailed:       l.metrics.messagesFailed,
		MessagesRetried:      l.metrics.messagesRetried,
		MessagesDeadLettered: l.metrics.messagesDeadLettered,
		AvgProcessingTime:    avgProcessingTime.String(),
		ErrorsByType:         errorsByType,
		SuccessRate:          l.calculateSuccessRate(),
	}
}

func (l *Logger) calculateSuccessRate() float64 {
	total := l.metrics.messagesProcessed
	if total == 0 {
		return 0.0
	}
	return float64(l.metrics.messagesSucceeded) / float64(total) * 100
}

// MetricsSnapshot es un snapshot de las métricas
type MetricsSnapshot struct {
	Service              string            `json:"service"`
	Uptime               string            `json:"uptime"`
	MessagesProcessed    int64             `json:"messages_processed"`
	MessagesSucceeded    int64             `json:"messages_succeeded"`
	MessagesFailed       int64             `json:"messages_failed"`
	MessagesRetried      int64             `json:"messages_retried"`
	MessagesDeadLettered int64             `json:"messages_dead_lettered"`
	AvgProcessingTime    string            `json:"avg_processing_time"`
	ErrorsByType         map[string]int64  `json:"errors_by_type"`
	SuccessRate          float64           `json:"success_rate_percent"`
}

// LogMetrics imprime métricas actuales
func (l *Logger) LogMetrics() {
	metrics := l.GetMetrics()
	l.WithFields(logrus.Fields{
		"uptime":                 metrics.Uptime,
		"messages_processed":     metrics.MessagesProcessed,
		"messages_succeeded":     metrics.MessagesSucceeded,
		"messages_failed":        metrics.MessagesFailed,
		"messages_retried":       metrics.MessagesRetried,
		"messages_dead_lettered": metrics.MessagesDeadLettered,
		"avg_processing_time":    metrics.AvgProcessingTime,
		"success_rate":           metrics.SuccessRate,
		"errors_by_type":         metrics.ErrorsByType,
	}).Info("Metrics snapshot")
}
