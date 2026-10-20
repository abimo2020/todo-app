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
	dlxName   = "todo.dlx"
	dlqName   = "todo.created.dlq"
)

type Consumer struct {
	ch *amqp.Channel
}

func New(conn *amqp.Connection) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	err = declareTopology(ch)
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
	if err := json.Unmarshal(msg.Body, &event); err != nil || event.ID == "" || event.Title == "" {
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

func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(dlxName, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(dlqName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(dlqName, queueName, dlxName, false, nil); err != nil {
		return err
	}
	_, err := ch.QueueDeclare(queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    dlxName,
		"x-dead-letter-routing-key": queueName,
	})
	return err
}
