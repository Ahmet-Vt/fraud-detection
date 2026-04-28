package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"fraud-detection/internal/transaction"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	client *Client
}

func NewPublisher(client *Client) *Publisher {
	return &Publisher{client: client}
}

type TransactionCreatedMessage struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	Amount        int64     `json:"amount"`
	Location      string    `json:"location"`
	CreatedAt     time.Time `json:"created_at"`
}

func (p *Publisher) PublishTransactionCreated(ctx context.Context, evt transaction.TransactionCreatedEvent) error {
	if p == nil || p.client == nil {
		return fmt.Errorf("publisher is nil")
	}

	ch, err := p.client.OpenChannel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	message := TransactionCreatedMessage{
		TransactionID: evt.TransactionID,
		UserID:        evt.UserID,
		Amount:        evt.Amount,
		Location:      evt.Location,
		CreatedAt:     evt.CreatedAt,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal transaction created message: %w", err)
	}

	if err := ch.PublishWithContext(
		ctx,
		FraudDetectionExchange,
		TransactionCreatedKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	); err != nil {
		return fmt.Errorf("publish transaction created message: %w", err)
	}

	return nil
}
