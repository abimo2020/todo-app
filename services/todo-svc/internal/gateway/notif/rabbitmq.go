package notif

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/abimo2020/todo-svc/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queueName = "todo.created"
)

type RabbitMQNotifier struct {
	ch *amqp.Channel
	mu sync.Mutex
}

func NewRabbitMQ(conn *amqp.Connection) (*RabbitMQNotifier, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return nil, err
	}

	return &RabbitMQNotifier{ch: ch}, nil
}

func (n *RabbitMQNotifier) NotifyTodoCreated(todo model.Todo) error {
	body, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	n.mu.Lock()
	defer n.mu.Unlock()

	return n.ch.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}
