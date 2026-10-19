package todo

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/abimo2020/todo-svc/apperror"
	"github.com/abimo2020/todo-svc/internal/model"
)

const maxTitleLength = 255

var (
	ErrEmptyTitle   = fmt.Errorf("%w: title cannot be empty", apperror.ErrInvalidInput)
	ErrTitleTooLong = fmt.Errorf("%w: title must be at most %d characters", apperror.ErrInvalidInput, maxTitleLength)
)

type TodoRepository interface {
	GetList() ([]model.Todo, error)
	GetDetail(id string) (model.Todo, error)
	Create(model.Todo) error
	Update(model.Todo) error
	Delete(id string) error
}

type NotifClient interface {
	NotifyTodoCreated(model.Todo) error
}
type Service struct {
	todoRepo    TodoRepository
	notifClient NotifClient
}

func New(todoRepo TodoRepository, notifClient NotifClient) *Service {
	return &Service{
		todoRepo:    todoRepo,
		notifClient: notifClient,
	}
}

func (s *Service) GetList() ([]model.Todo, error) {
	return s.todoRepo.GetList()
}

func (s *Service) GetDetail(id string) (model.Todo, error) {
	return s.todoRepo.GetDetail(id)
}

func (s *Service) Create(param model.Todo) (model.Todo, error) {
	title := strings.TrimSpace(param.Title)
	if title == "" {
		return model.Todo{}, ErrEmptyTitle
	}
	if utf8.RuneCountInString(title) > maxTitleLength {
		return model.Todo{}, ErrTitleTooLong
	}

	param.Title = title
	param.ID = generateID()
	param.CreatedAt = time.Now()

	err := s.todoRepo.Create(param)
	if err != nil {
		return model.Todo{}, err
	}

	err = s.notifClient.NotifyTodoCreated(param)
	if err != nil {
		slog.Warn("failed to send todo created notification",
			"todo_id", param.ID,
			"error", err,
		)
	}

	return param, nil
}

func (s *Service) MarkDone(id string) error {
	todo, err := s.todoRepo.GetDetail(id)
	if err != nil {
		return err
	}
	todo.Done = true
	return s.todoRepo.Update(todo)
}

func (s *Service) Delete(id string) error {
	return s.todoRepo.Delete(id)
}

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
