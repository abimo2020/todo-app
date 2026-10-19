package todocreated

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/abimo2020/notif-svc/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queueName = "todo.created"
)

type Consumer struct {
	ch *amqp.Channel
}

func New(conn *amqp.Connection) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &Consumer{ch: ch}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	msgs, err := c.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	slog.Info("active consumer", "queue", queueName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("Consumer channel was closed")
			}
			c.handle(msg)
		}
	}
}

func (c *Consumer) handle(msg amqp.Delivery) {
	var event model.TodoCreatedEvent
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		slog.Warn(
			"invalid message",
			"queue", queueName,
			"error", err,
			"body", string(msg.Body),
		)
		msg.Nack(false, false)
		return
	}

	slog.Info("new todo created",
		"todo_id", event.ID,
		"title", event.Title,
	)
	msg.Ack(false)
}
