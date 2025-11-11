package common

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const (
	MaxRetries        = 3
	RetryCountHeader  = "x-retry-count"
	RetryReasonHeader = "x-retry-reason"
)

// RetryHandler maneja la lógica de reintentos y DLQ
type RetryHandler struct {
	channel     *amqp.Channel
	queueName   string
	dlqName     string
	dlxName     string
	maxRetries  int
}

// NewRetryHandler crea un nuevo handler de reintentos
func NewRetryHandler(channel *amqp.Channel, queueName string, maxRetries int) (*RetryHandler, error) {
	if maxRetries <= 0 {
		maxRetries = MaxRetries
	}

	dlqName := queueName + "_dlq"
	dlxName := queueName + "_dlx"

	handler := &RetryHandler{
		channel:    channel,
		queueName:  queueName,
		dlqName:    dlqName,
		dlxName:    dlxName,
		maxRetries: maxRetries,
	}

	// Configurar DLX y DLQ
	if err := handler.setupDLQ(); err != nil {
		return nil, err
	}

	return handler, nil
}

// setupDLQ configura el Dead Letter Exchange y Queue
func (rh *RetryHandler) setupDLQ() error {
	// 1. Declarar Dead Letter Exchange
	err := rh.channel.ExchangeDeclare(
		rh.dlxName,  // name
		"fanout",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX: %v", err)
	}

	// 2. Declarar Dead Letter Queue
	_, err = rh.channel.QueueDeclare(
		rh.dlqName, // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %v", err)
	}

	// 3. Bind DLQ al DLX
	err = rh.channel.QueueBind(
		rh.dlqName, // queue name
		"",         // routing key
		rh.dlxName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ to DLX: %v", err)
	}

	// 4. Configurar la cola principal con DLX
	args := amqp.Table{
		"x-dead-letter-exchange": rh.dlxName,
	}

	_, err = rh.channel.QueueDeclare(
		rh.queueName, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		args,         // arguments (con DLX)
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue with DLX: %v", err)
	}

	log.Printf("DLQ setup complete: queue=%s, dlq=%s, dlx=%s", rh.queueName, rh.dlqName, rh.dlxName)
	return nil
}

// HandleError maneja un mensaje fallido con lógica de reintentos
// Retorna true si el mensaje fue manejado (ack/nack hecho), false si debe ser manejado manualmente
func (rh *RetryHandler) HandleError(msg amqp.Delivery, err error, shouldRetry bool) bool {
	retryCount := rh.getRetryCount(msg)

	log.Printf("Message failed: retry_count=%d, error=%v", retryCount, err)

	// Si no debe reintentar (mensaje malformado), enviar directamente a DLQ
	if !shouldRetry {
		log.Printf("Non-retryable error, sending to DLQ")
		return rh.sendToDLQ(msg, err, "non_retryable_error")
	}

	// Si ya alcanzó el máximo de reintentos, enviar a DLQ
	if retryCount >= rh.maxRetries {
		log.Printf("Max retries reached (%d), sending to DLQ", rh.maxRetries)
		return rh.sendToDLQ(msg, err, "max_retries_exceeded")
	}

	// Reintentar: incrementar contador y reencolar
	log.Printf("Requeuing message (retry %d/%d)", retryCount+1, rh.maxRetries)
	return rh.requeueWithRetry(msg, err, retryCount+1)
}

// getRetryCount obtiene el contador de reintentos del header
func (rh *RetryHandler) getRetryCount(msg amqp.Delivery) int {
	if msg.Headers == nil {
		return 0
	}

	if count, ok := msg.Headers[RetryCountHeader].(int32); ok {
		return int(count)
	}

	if count, ok := msg.Headers[RetryCountHeader].(int64); ok {
		return int(count)
	}

	return 0
}

// requeueWithRetry reencola el mensaje con contador incrementado
func (rh *RetryHandler) requeueWithRetry(msg amqp.Delivery, err error, newRetryCount int) bool {
	// Preparar headers con contador actualizado
	headers := amqp.Table{}
	if msg.Headers != nil {
		for k, v := range msg.Headers {
			headers[k] = v
		}
	}
	headers[RetryCountHeader] = int32(newRetryCount)
	headers[RetryReasonHeader] = err.Error()

	// Publicar de nuevo a la cola
	publishErr := rh.channel.Publish(
		"",           // exchange (default)
		rh.queueName, // routing key (nombre de la cola)
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  msg.ContentType,
			Body:         msg.Body,
			Headers:      headers,
			DeliveryMode: amqp.Persistent, // persist messages
		},
	)

	if publishErr != nil {
		log.Printf("❌ Failed to requeue message: %v", publishErr)
		// Nack sin requeue para que RabbitMQ lo maneje
		msg.Nack(false, false)
		return false
	}

	// Ack el mensaje original (ya lo republicamos)
	msg.Ack(false)
	return true
}

// sendToDLQ envía el mensaje a la Dead Letter Queue
func (rh *RetryHandler) sendToDLQ(msg amqp.Delivery, err error, reason string) bool {
	// Preparar headers con información del error
	headers := amqp.Table{}
	if msg.Headers != nil {
		for k, v := range msg.Headers {
			headers[k] = v
		}
	}
	headers[RetryCountHeader] = int32(rh.getRetryCount(msg))
	headers[RetryReasonHeader] = err.Error()
	headers["x-dlq-reason"] = reason

	// Publicar directamente a la DLQ
	publishErr := rh.channel.Publish(
		rh.dlxName, // exchange (DLX)
		"",         // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  msg.ContentType,
			Body:         msg.Body,
			Headers:      headers,
			DeliveryMode: amqp.Persistent,
		},
	)

	if publishErr != nil {
		log.Printf("Failed to send to DLQ: %v", publishErr)
		// Nack sin requeue
		msg.Nack(false, false)
		return false
	}

	log.Printf("Message sent to DLQ: %s", rh.dlqName)

	// Ack el mensaje original
	msg.Ack(false)
	return true
}

// AckSuccess marca un mensaje como procesado exitosamente
func (rh *RetryHandler) AckSuccess(msg amqp.Delivery) {
	msg.Ack(false)
	log.Printf("Message processed successfully")
}
