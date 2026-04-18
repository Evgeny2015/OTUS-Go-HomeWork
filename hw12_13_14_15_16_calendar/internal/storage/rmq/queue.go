package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Queue defines the interface for message queue operations
type Queue interface {
	Connect(ctx context.Context) error
	Close() error
	PublishNotification(ctx context.Context, notification *internal.Notification) error
	ConsumeNotifications(ctx context.Context) (<-chan internal.Notification, error)
}

// RabbitMQ implements Queue interface using RabbitMQ
type RabbitMQ struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	queueName    string
	exchangeName string
	uri          string
}

// NewRabbitMQ creates a new RabbitMQ instance
func NewRabbitMQ(uri, queueName, exchangeName string) *RabbitMQ {
	return &RabbitMQ{
		uri:          uri,
		queueName:    queueName,
		exchangeName: exchangeName,
	}
}

// Connect establishes connection to RabbitMQ and creates necessary structures
func (r *RabbitMQ) Connect(ctx context.Context) error {
	var err error
	r.conn, err = amqp.Dial(r.uri)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare exchange
	err = r.channel.ExchangeDeclare(
		r.exchangeName, // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = r.channel.QueueDeclare(
		r.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = r.channel.QueueBind(
		r.queueName,     // queue name
		"notifications", // routing key
		r.exchangeName,  // exchange
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Printf("Connected to RabbitMQ at %s, queue: %s, exchange: %s", r.uri, r.queueName, r.exchangeName)
	return nil
}

// Close closes the connection and channel
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// PublishNotification publishes a notification to the queue
func (r *RabbitMQ) PublishNotification(ctx context.Context, notification *internal.Notification) error {
	if r.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	body, err := notification.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Set message expiration to 1 year (in milliseconds) for old notifications cleanup
	expiration := "31536000000" // 365 days in milliseconds

	err = r.channel.PublishWithContext(
		ctx,
		r.exchangeName,  // exchange
		"notifications", // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Expiration:   expiration,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	log.Printf("Published notification for event %s to user %s", notification.EventID, notification.UserID)
	return nil
}

// ConsumeNotifications starts consuming notifications from the queue
func (r *RabbitMQ) ConsumeNotifications(ctx context.Context) (<-chan internal.Notification, error) {
	if r.channel == nil {
		return nil, fmt.Errorf("channel is not initialized")
	}

	msgs, err := r.channel.Consume(
		r.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}

	notifications := make(chan internal.Notification)

	go func() {
		defer close(notifications)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				var notification internal.Notification
				if err := json.Unmarshal(msg.Body, &notification); err != nil {
					log.Printf("Failed to unmarshal notification: %v", err)
					msg.Nack(false, false) // reject message without requeue
					continue
				}
				notifications <- notification
				msg.Ack(false)
			}
		}
	}()

	return notifications, nil
}
