package notif

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/abimo2020/todo-svc/internal/model"
)

type Notification struct {
	baseURL string
	client  http.Client
}

func New(baseURL string) *Notification {
	return &Notification{baseURL: baseURL}
}

func (n *Notification) NotifyTodoCreated(todo model.Todo) error {
	body, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	resp, err := n.client.Post(n.baseURL+"/notify", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notif-svc returned status %d", resp.StatusCode)
	}
	return nil
}
