package todo

import (
	"errors"
	"sync"

	"github.com/abimo2020/todo-svc/internal/model"
)

var (
	ErrNotFound = errors.New("todo not found")
)

type Interface interface {
	GetList() ([]model.Todo, error)
	GetDetail(id string) (model.Todo, error)
	Create(todo model.Todo) error
	Update(todo model.Todo) error
	Delete(id string) error
}

type todo struct {
	mu    sync.Mutex
	todos map[string]model.Todo
}

func Init() Interface {
	return &todo{
		todos: make(map[string]model.Todo),
	}
}

func (t *todo) GetList() ([]model.Todo, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	result := make([]model.Todo, 0, len(t.todos))
	for _, v := range t.todos {
		result = append(result, v)
	}
	return result, nil
}
func (t *todo) GetDetail(id string) (model.Todo, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	result, ok := t.todos[id]
	if !ok {
		return model.Todo{}, ErrNotFound
	}
	return result, nil
}
func (t *todo) Create(todo model.Todo) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.todos[todo.ID] = todo
	return nil
}
func (t *todo) Update(todo model.Todo) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, ok := t.todos[todo.ID]
	if !ok {
		return ErrNotFound
	}

	t.todos[todo.ID] = todo
	return nil
}
func (t *todo) Delete(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, ok := t.todos[id]
	if !ok {
		return ErrNotFound
	}

	delete(t.todos, id)
	return nil
}
