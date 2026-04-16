package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
}

func New(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) Connection() *amqp.Connection {
	if c == nil {
		return nil
	}

	return c.conn
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
