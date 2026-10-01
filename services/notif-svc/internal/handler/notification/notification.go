package notification

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/abimo2020/notif-svc/internal/model"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /notify", h.notify)
}

func (h *Handler) notify(w http.ResponseWriter, r *http.Request) {
	var event model.TodoCreatedEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("[NOTIF] Todo baru dibuat -> ID: %s, Title: %s\n", event.ID, event.Title)
	w.WriteHeader(http.StatusOK)
}
