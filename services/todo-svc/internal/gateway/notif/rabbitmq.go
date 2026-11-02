package notif

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/abimo2020/todo-svc/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queueName = "todo.created"
	dlxName   = "todo.dlx"
	dlqName   = "todo.created.dlq"

	dialTimeout   = 2 * time.Second
	retryCooldown = 5 * time.Second
)

var (
	errBrokerUnavailable = errors.New("rabbitmq unavailable, waiting before reconnect.")
)

type RabbitMQNotifier struct {
	url string

	conn       *amqp.Connection
	ch         *amqp.Channel
	mu         sync.Mutex
	retryAfter time.Time
}

func NewRabbitMQ(url string) *RabbitMQNotifier {
	// ch, err := conn.Channel()
	// if err != nil {
	// 	return nil, err
	// }

	// if err := declareTopology(ch); err != nil {
	// 	return nil, err
	// }

	// if err := ch.Confirm(false); err != nil {
	// 	return nil, err
	// }

	return &RabbitMQNotifier{url: url}
}

func (n *RabbitMQNotifier) NotifyTodoCreated(todo model.Todo) error {
	body, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	confirmation, err := n.publish(ctx, body)
	if err != nil {
		return err
	}

	acked, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !acked {
		return errors.New("broker rejected message (nack)")
	}

	return nil
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

func (n *RabbitMQNotifier) publish(ctx context.Context, body []byte) (*amqp.DeferredConfirmation, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	ch, err := n.ensureChannel()
	if err != nil {
		return nil, err
	}

	confirmation, err := ch.PublishWithDeferredConfirmWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	if err != nil {
		n.closedLocked()
		return nil, err
	}
	return confirmation, nil
}

func (n *RabbitMQNotifier) ensureChannel() (*amqp.Channel, error) {
	if n.ch != nil && !n.ch.IsClosed() {
		return n.ch, nil
	}
	n.closedLocked()

	if time.Now().Before(n.retryAfter) {
		return nil, errBrokerUnavailable
	}
	conn, ch, err := n.connect()
	if err != nil {
		n.retryAfter = time.Now().Add(retryCooldown)
		return nil, err
	}

	n.conn, n.ch = conn, ch
	return ch, nil
}

func (n *RabbitMQNotifier) connect() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.DialConfig(n.url, amqp.Config{
		Dial: amqp.DefaultDial(dialTimeout),
	})
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	if err := declareTopology(ch); err != nil {
		conn.Close()
		return nil, nil, err
	}
	if err := ch.Confirm(false); err != nil {
		conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

func (n *RabbitMQNotifier) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.closedLocked()
}

func (n *RabbitMQNotifier) closedLocked() error {
	var err error
	if n.conn != nil && !n.conn.IsClosed() {
		err = n.conn.Close()
	}
	n.conn, n.ch = nil, nil
	return err
}
