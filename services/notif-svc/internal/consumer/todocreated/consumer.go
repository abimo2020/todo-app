package todocreated

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

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
		fmt.Printf("[NOTIF] invalid message: %v", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("[NOTIF] New todo is created -> ID: %s, Title: %s", event.ID, event.Title)
	msg.Ack(false)
}
