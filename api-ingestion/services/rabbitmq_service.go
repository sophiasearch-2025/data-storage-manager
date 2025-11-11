package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQService(url string) (*RabbitMQService, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// NOTA: La cola ingestion_queue es creada por el worker-indexer con DLX configurado
	// La API solo publica mensajes, no necesita declarar la cola

	return &RabbitMQService{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQService) PublishNews(news interface{}) error {
	body, err := json.Marshal(news)
	if err != nil {
		return fmt.Errorf("failed to marshal news: %v", err)
	}

	err = r.channel.Publish("", "ingestion_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf("Published message to queue")
	return nil
}

func (r *RabbitMQService) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
