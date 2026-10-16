package todo

import (
	"time"

	"github.com/abimo2020/todo-svc/internal/model"
)

type createRequest struct {
	Title string `json:"title"`
}

type todoResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(t model.Todo) todoResponse {
	return todoResponse{
		ID:        t.ID,
		Title:     t.Title,
		Done:      t.Done,
		CreatedAt: t.CreatedAt,
		UpdatedAt: *t.UpdatedAt,
	}
}
