package rabbitmq

import (
	"fmt"
)

const (
	FraudDetectionExchange = "fraud.detection"
	FraudTransactionQueue  = "fraud.transaction.queue"
	TransactionCreatedKey  = "transaction.created"
)

func (c *Client) DeclareTopology() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		FraudDetectionExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare fraud detection exchange: %w", err)
	}

	if _, err := ch.QueueDeclare(
		FraudTransactionQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare fraud detection queue: %w", err)
	}

	if err := ch.QueueBind(
		FraudTransactionQueue,
		TransactionCreatedKey,
		FraudDetectionExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind fraud detection queue: %w", err)
	}

	return nil
}

