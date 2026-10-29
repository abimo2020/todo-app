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

	maxBackoff     = 30 * time.Second
	dialTimeout    = 5 * time.Second
	stableDuration = 30 * time.Second
)

type Consumer struct {
	url string
}

func New(url string) *Consumer {
	return &Consumer{url: url}
}

func (c *Consumer) Run(ctx context.Context) error {
	backoff := time.Second

	for {
		var readyAt time.Time

		err := c.session(ctx, func() { readyAt = time.Now() })

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !readyAt.IsZero() && time.Since(readyAt) > stableDuration {
			backoff = time.Second
		}

		slog.Warn("RabbitMQ disconnected, try to reconnecting",
			"error", err,
			"retry_in", backoff,
		)

		select {
		case <-ctx.Done():
			return err
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func (c *Consumer) handle(ch *amqp.Channel, msg amqp.Delivery) {
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
		c.retryOrDeadLetter(ch, msg, err)
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

func (n *Consumer) retryOrDeadLetter(ch *amqp.Channel, msg amqp.Delivery, cause error) {
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

	err := ch.PublishWithContext(ctx, "", retryQueueName, false, false, amqp.Publishing{
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

func (c *Consumer) session(ctx context.Context, onReady func()) error {
	conn, err := amqp.DialConfig(c.url, amqp.Config{
		Dial: amqp.DefaultDial(dialTimeout),
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := declareTopology(ch); err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	slog.Info("consumer aktif", "queue", queueName)

	onReady()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			{
				if !ok {
					return errors.New("connection or channel was closed")
				}

				c.handle(ch, msg)
			}
		}
	}
}
