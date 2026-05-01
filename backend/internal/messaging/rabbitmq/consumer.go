package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TransactionCreatedHandler func(ctx context.Context, evt TransactionCreatedMessage) error

type Consumer struct {
	client   *Client
	handle   TransactionCreatedHandler
	tag      string
	prefetch int
}

func NewConsumer(client *Client, handle TransactionCreatedHandler, tag string, prefetch int) *Consumer {
	if tag == "" {
		tag = "fraud-worker"
	}
	if prefetch <= 0 {
		prefetch = 1
	}
	return &Consumer{client: client, handle: handle, tag: tag, prefetch: prefetch}
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

	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return fmt.Errorf("consumer qos: %w", err)
	}

	deliveries, err := ch.Consume(
		FraudTransactionQueue,
		c.tag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consumer consume: %w", err)
	}
	log.Printf("consumer: started queue=%s tag=%s prefetch=%d", FraudTransactionQueue, c.tag, c.prefetch)

	for {
		select {
		case <-ctx.Done():
			log.Printf("consumer: stopping queue=%s reason=%v", FraudTransactionQueue, ctx.Err())
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("consumer: delivery channel closed")
			}
			log.Printf(
				"consumer: received queue=%s routing_key=%s message_id=%s size=%d",
				FraudTransactionQueue,
				d.RoutingKey,
				d.MessageId,
				len(d.Body),
			)
			c.dispatch(ctx, d)
		}
	}
}

func (c *Consumer) dispatch(ctx context.Context, d amqp.Delivery) {
	var evt TransactionCreatedMessage
	if err := json.Unmarshal(d.Body, &evt); err != nil {
		log.Printf("consumer: invalid json message_id=%s: %v", d.MessageId, err)
		if err := d.Nack(false, false); err != nil {
			log.Printf("consumer: nack: %v", err)
		}
		log.Printf("consumer: nacked message_id=%s requeue=false reason=invalid_json", d.MessageId)
		return
	}

	log.Printf("consumer: processing transaction_id=%s user_id=%s message_id=%s", evt.TransactionID, evt.UserID, d.MessageId)
	if err := c.handle(ctx, evt); err != nil {
		log.Printf("consumer: handler error message_id=%s transaction_id=%s: %v", d.MessageId, evt.TransactionID, err)
		if err := d.Nack(false, false); err != nil {
			log.Printf("consumer: nack: %v", err)
		}
		log.Printf("consumer: nacked message_id=%s requeue=false reason=handler_error", d.MessageId)
		return
	}

	if err := d.Ack(false); err != nil {
		log.Printf("consumer: ack: %v", err)
		return
	}
	log.Printf("consumer: acked message_id=%s transaction_id=%s", d.MessageId, evt.TransactionID)
}
