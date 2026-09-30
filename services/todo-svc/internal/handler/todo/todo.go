package todo

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/abimo2020/todo-svc/internal/model"
)

type TodoService interface {
	GetList() ([]model.Todo, error)
	GetDetail(id string) (model.Todo, error)
	Create(param model.Todo) (model.Todo, error)
	MarkDone(id string) error
	Delete(id string) error
}

type Handler struct {
	todoService TodoService
}

func New(todoService TodoService) *Handler {
	return &Handler{
		todoService: todoService,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /todos", h.list)
	mux.HandleFunc("GET /todos/{id}", h.detail)
	mux.HandleFunc("POST /todos", h.create)
	mux.HandleFunc("PATCH /todos/{id}/done", h.markDone)
	mux.HandleFunc("DEL /todos/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.todoService.GetList()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	res, err := h.todoService.GetDetail(id)
	if err != nil {
		if errors.Is(err, err) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var data model.Todo
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeError(w, http.StatusInternalServerError, err)
	}
	res, err := h.todoService.Create(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) markDone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.todoService.MarkDone(id); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.todoService.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}
