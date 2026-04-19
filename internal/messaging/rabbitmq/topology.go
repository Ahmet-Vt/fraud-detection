package rabbitmq

import (
	"fmt"
)

func (c *Client) DeclareTopology() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		"fraud.detection",
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
		"fraud.transaction.queue",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare fraud detection queue: %w", err)
	}

	if err := ch.QueueBind(
		"fraud.transaction.queue",
		"transaction.created",
		"fraud.detection",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind fraud detection queue: %w", err)
	}


	

	return nil
}

