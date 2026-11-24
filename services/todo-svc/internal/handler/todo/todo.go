package todo

import (
	"fmt"
	"net/http"

	"github.com/abimo2020/todo-svc/apperror"
	"github.com/abimo2020/todo-svc/internal/model"
	"github.com/gin-gonic/gin"
)

const maxBodyBytes = 1 << 20 // 1 MiB

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

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/todos", h.list)
	r.GET("/todos/:id", h.detail)
	r.POST("/todos", h.create)
	r.PATCH("/todos/:id/done", h.markDone)
	r.DELETE("/todos/:id", h.delete)
}

func (h *Handler) list(c *gin.Context) {
	res, err := h.todoService.GetList()
	if err != nil {
		responseError(c, err)
		return
	}

	out := make([]todoResponse, 0, len(res))
	for _, v := range res {
		out = append(out, toResponse(v))
	}

	c.JSON(http.StatusOK, out)
}

func (h *Handler) detail(c *gin.Context) {
	id := c.Param("id")

	res, err := h.todoService.GetDetail(id)
	if err != nil {
		responseError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(res))
}

func (h *Handler) create(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responseError(c, fmt.Errorf("%w: %v", apperror.ErrInvalidInput, err))
		return
	}
	res, err := h.todoService.Create(model.Todo{Title: req.Title})
	if err != nil {
		responseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toResponse(res))
}

func (h *Handler) markDone(c *gin.Context) {
	id := c.Param("id")

	if err := h.todoService.MarkDone(id); err != nil {
		responseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.todoService.Delete(id); err != nil {
		responseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func responseError(c *gin.Context, err error) {
	c.JSON(apperror.HTTPStatus(err), gin.H{"error": err})
}
