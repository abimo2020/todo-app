package notification

import (
	"context"
	"net/http"
	"strconv"

	"github.com/abimo2020/notif-svc/internal/model"
	"github.com/gin-gonic/gin"
)

type Reader interface {
	List(ctx context.Context, limit int64) ([]model.Notification, error)
}

type Handler struct {
	reader Reader
}

func New(reader Reader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/notifications", h.list)
}

func (h *Handler) list(c *gin.Context) {
	limit := int64(20)
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 1 || v > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit should be numeric between 1 > 100"})
			return
		}
		limit = v
	}

	notifList, err := h.reader.List(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notification"})
		return
	}

	out := make([]notificationResponse, 0, len(notifList))
	for _, v := range notifList {
		out = append(out, toResponse(v))
	}
	c.JSON(http.StatusOK, out)
}
