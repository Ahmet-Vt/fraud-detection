package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fraud-detection/internal/transaction"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	client *Client
	mu     sync.Mutex
	ch     *amqp.Channel
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

	p.mu.Lock()
	defer p.mu.Unlock()

	ch, err := p.ensureChannel()
	if err != nil {
		return err
	}

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

func (p *Publisher) ensureChannel() (*amqp.Channel, error) {
	if p.ch != nil && !p.ch.IsClosed() {
		return p.ch, nil
	}

	ch, err := p.client.OpenChannel()
	if err != nil {
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}
	p.ch = ch

	return p.ch, nil
}

func (p *Publisher) Close() error {
	if p == nil {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch == nil {
		return nil
	}

	err := p.ch.Close()
	p.ch = nil
	return err
}
