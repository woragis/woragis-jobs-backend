package database

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConnection manages RabbitMQ connection and channel
type RabbitMQConnection struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ(url string) (*RabbitMQConnection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQConnection{
		Connection: conn,
		Channel:    ch,
	}, nil
}

// Close closes the RabbitMQ connection and channel
func (r *RabbitMQConnection) Close() error {
	if r.Channel != nil {
		if err := r.Channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	if r.Connection != nil {
		if err := r.Connection.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	return nil
}
