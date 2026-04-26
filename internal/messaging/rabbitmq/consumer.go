package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const workerTag = "fraud-worker"

type TransactionCreatedHandler func(ctx context.Context, evt TransactionCreatedMessage) error

type Consumer struct {
	client *Client
	handle TransactionCreatedHandler
}

func NewConsumer(client *Client, handle TransactionCreatedHandler) *Consumer {
	return &Consumer{client: client, handle: handle}
}


func (c *Consumer) Run(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("consumer: nil client")
	}
	if c.handle == nil {
		return fmt.Errorf("consumer: nil handler")
	}

	ch, err := c.client.OpenChannel()
	if err != nil {
		return fmt.Errorf("consumer open channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("consumer qos: %w", err)
	}

	deliveries, err := ch.Consume(
		FraudTransactionQueue,
		workerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consumer consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("consumer: delivery channel closed")
			}
			c.dispatch(ctx, d)
		}
	}
}

func (c *Consumer) dispatch(ctx context.Context, d amqp.Delivery) {
	var evt TransactionCreatedMessage
	if err := json.Unmarshal(d.Body, &evt); err != nil {
		log.Printf("consumer: invalid json: %v", err)
		if err := d.Nack(false, false); err != nil {
			log.Printf("consumer: nack: %v", err)
		}
		return
	}

	if err := c.handle(ctx, evt); err != nil {
		log.Printf("consumer: handler error: %v", err)
		if err := d.Nack(false, false); err != nil {
			log.Printf("consumer: nack: %v", err)
		}
		return
	}

	if err := d.Ack(false); err != nil {
		log.Printf("consumer: ack: %v", err)
	}
}
