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
	dlxName   = "todo.dlx"
	dlqName   = "todo.created.dlq"
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

	if err := declareTopology(ch); err != nil {
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
