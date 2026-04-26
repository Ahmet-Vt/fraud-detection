package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"fraud-detection/internal/domain"
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

func (p *Publisher) PublishTransactionCreated(ctx context.Context, transaction *domain.Transaction) error {
	if p == nil || p.client == nil {
		return fmt.Errorf("publisher is nil")
	}
	if transaction == nil {
		return fmt.Errorf("transaction is nil")
	}

	ch, err := p.client.OpenChannel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	message := TransactionCreatedMessage{
		TransactionID: transaction.ID.String(),
		UserID:        transaction.UserID.String(),
		Amount:        transaction.Amount,
		Location:      transaction.Location,
		CreatedAt:     transaction.CreatedAt,
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
