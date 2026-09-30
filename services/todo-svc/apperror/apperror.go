package apperror

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrEmptyTitle   = errors.New("title cannot be empty")
	ErrInvalidInput = errors.New("invalid input")
)

func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrEmptyTitle), errors.Is(err, ErrInvalidInput):
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
