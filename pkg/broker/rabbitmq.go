package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQBroker handles the connection and publishing to RabbitMQ.
type RabbitMQBroker struct {
	conn *amqp091.Connection
}

// NewRabbitMQBroker creates and returns a new RabbitMQBroker.
func NewRabbitMQBroker(url string) (*RabbitMQBroker, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &RabbitMQBroker{conn: conn}, nil
}

// Publish sends a message to a specific queue.
// The body is marshaled to JSON.
func (b *RabbitMQBroker) Publish(ctx context.Context, queueName string, body interface{}) error {
	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Declare a durable queue, so it survives broker restarts.
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Marshal the body to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body to JSON: %w", err)
	}

	// Publish the message
	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key (queue name)
		false,  // mandatory
		false,  // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent, // Make message persistent
			Timestamp:    time.Now(),
			Body:         jsonBody,
		})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	log.Printf("Successfully published message to queue: %s", queueName)
	return nil
}

func (b *RabbitMQBroker) Consume(queueName string, handler func(d amqp091.Delivery)) error {
	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	// Do not defer ch.Close() here, as the consumer runs indefinitely.

	// Declare the queue to ensure it exists.
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Start consuming messages.
	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer tag (let RabbitMQ generate one)
		false,     // auto-ack (we want to manually ack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Run the message handling in a separate goroutine.
	go func() {
		for d := range msgs {
			handler(d)
		}
	}()

	log.Printf("Consumer started for queue: %s. Waiting for messages.", queueName)
	return nil
}

// Close closes the connection to RabbitMQ.
func (b *RabbitMQBroker) Close() {
	if b.conn != nil {
		b.conn.Close()
	}
}
