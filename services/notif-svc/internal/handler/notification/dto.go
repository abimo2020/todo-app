package notification

import (
	"time"

	"github.com/abimo2020/notif-svc/internal/model"
)

type notificationResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	TodoID    string    `json:"todo_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

func toResponse(in model.Notification) notificationResponse {
	return notificationResponse{
		ID:        in.ID,
		Type:      in.Type,
		TodoID:    in.TodoID,
		Title:     in.Title,
		CreatedAt: in.CreatedAt,
	}
}
