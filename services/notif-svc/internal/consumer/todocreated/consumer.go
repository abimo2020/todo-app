package todocreated

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/abimo2020/notif-svc/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queueName      = "todo.created"
	dlxName        = "todo.dlx"
	dlqName        = "todo.created.dlq"
	retryQueueName = "todo.created.retry"

	maxRetries = 3
	retryDelay = 10 * time.Second
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

	if err := c.process(event); err != nil {
		c.retryOrDeadLetter(msg, err)
		return
	}
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
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    dlxName,
		"x-dead-letter-routing-key": queueName,
	}); err != nil {
		return err
	}

	_, err := ch.QueueDeclare(retryQueueName, true, false, false, false, amqp.Table{
		"x-message-ttl":             int32(retryDelay.Milliseconds()),
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": queueName,
	})
	return err
}

func (n *Consumer) process(event model.TodoCreatedEvent) error {
	slog.Info("new todo created",
		"todo_id", event.ID,
		"title", event.Title,
	)
	return nil
}

func (n *Consumer) retryOrDeadLetter(msg amqp.Delivery, cause error) {
	attempt := retryCount(msg.Headers)
	if attempt >= maxRetries {
		slog.Warn("reach limit retry, send into dead letter queue",
			"attempts", attempt+1,
			"error", cause,
			"body", string(msg.Body),
		)
		msg.Nack(false, false)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := n.ch.PublishWithContext(ctx, "", retryQueueName, false, false, amqp.Publishing{
		Headers:      amqp.Table{"x-retry-count": int32(attempt + 1)},
		Body:         msg.Body,
		ContentType:  msg.ContentType,
		DeliveryMode: amqp.Persistent,
	})
	if err != nil {
		slog.Error("failed to reschedule retry, message resend into queue", "error", err)
	}
	slog.Warn("failed to process, reschedule the retry",
		"attempt", attempt+1,
		"max_retries", maxRetries,
		"retry_in", retryDelay.String,
		"error", cause,
	)
	msg.Ack(false)
}

func retryCount(h amqp.Table) int {
	switch v := h["x-retry-count"].(type) {
	case int32:
		return int(v)
	case int64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}
