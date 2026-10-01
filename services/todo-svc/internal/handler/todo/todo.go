package todo

import (
	"encoding/json"
	"net/http"

	"github.com/abimo2020/todo-svc/apperror"
	"github.com/abimo2020/todo-svc/internal/model"
	"github.com/abimo2020/todo-svc/pkg/httputil"
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
		httpStatus := apperror.HTTPStatus(err)
		httputil.WriteError(w, httpStatus, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	res, err := h.todoService.GetDetail(id)
	if err != nil {
		httpStatus := apperror.HTTPStatus(err)
		httputil.WriteError(w, httpStatus, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var data model.Todo
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
	}
	res, err := h.todoService.Create(data)
	if err != nil {
		httpStatus := apperror.HTTPStatus(err)
		httputil.WriteError(w, httpStatus, err)
	}
	httputil.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) markDone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.todoService.MarkDone(id); err != nil {
		httpStatus := apperror.HTTPStatus(err)
		httputil.WriteError(w, httpStatus, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.todoService.Delete(id); err != nil {
		httpStatus := apperror.HTTPStatus(err)
		httputil.WriteError(w, httpStatus, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
