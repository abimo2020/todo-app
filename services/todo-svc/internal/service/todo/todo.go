package todo

import (
	"strconv"
	"time"

	"github.com/abimo2020/todo-svc/apperror"
	"github.com/abimo2020/todo-svc/internal/model"
)

type TodoRepository interface {
	GetList() ([]model.Todo, error)
	GetDetail(id string) (model.Todo, error)
	Create(model.Todo) error
	Update(model.Todo) error
	Delete(id string) error
}

type Service struct {
	todoRepo TodoRepository
}

func New(todoRepo TodoRepository) *Service {
	return &Service{
		todoRepo: todoRepo,
	}
}

func (s *Service) GetList() ([]model.Todo, error) {
	return s.todoRepo.GetList()
}

func (s *Service) GetDetail(id string) (model.Todo, error) {
	return s.todoRepo.GetDetail(id)
}

func (s *Service) Create(param model.Todo) (model.Todo, error) {
	if param.Title == "" {
		return model.Todo{}, apperror.ErrEmptyTitle
	}

	param.ID = generateID()
	param.CreatedAt = time.Now()

	err := s.todoRepo.Create(param)
	if err != nil {
		return model.Todo{}, err
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
