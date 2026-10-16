package todo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/abimo2020/todo-svc/apperror"
	"github.com/abimo2020/todo-svc/internal/model"
)

type fakeService struct {
	todo model.Todo
	err  error
}

func (f *fakeService) GetList() ([]model.Todo, error)              { return []model.Todo{f.todo}, f.err }
func (f *fakeService) GetDetail(id string) (model.Todo, error)     { return f.todo, f.err }
func (f *fakeService) MarkDone(id string) error                    { return f.err }
func (f *fakeService) Create(param model.Todo) (model.Todo, error) { return f.todo, f.err }
func (f *fakeService) Delete(id string) error                      { return f.err }

func TestHandler(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		svc        *fakeService
		wantStatus int
	}{
		{
			name:       "success to create",
			method:     http.MethodPost,
			path:       "/todos",
			body:       `{"title": "Bebek goreng"}`,
			svc:        &fakeService{todo: model.Todo{ID: "1", Title: "Bebek goreng"}},
			wantStatus: http.StatusCreated,
		}, {
			name:       "invalid json",
			method:     http.MethodPost,
			path:       "/todos",
			body:       `{json rusak`,
			svc:        &fakeService{},
			wantStatus: http.StatusBadRequest,
		}, {
			name:       "todo not found",
			method:     http.MethodGet,
			path:       "/todos/x1",
			svc:        &fakeService{err: apperror.ErrNotFound},
			wantStatus: http.StatusNotFound,
		}, {
			name:       "todo found",
			method:     http.MethodGet,
			path:       "/todos/1",
			svc:        &fakeService{todo: model.Todo{ID: "1", Title: "Bebek goreng", CreatedAt: now, UpdatedAt: &now}},
			wantStatus: http.StatusOK,
		}, {
			name:       "unknown fields are rejected",
			method:     http.MethodPost,
			path:       "/todos",
			svc:        &fakeService{err: apperror.ErrInvalidInput},
			body:       `{"title": "salah", "done": true}`,
			wantStatus: http.StatusBadRequest,
		}, {
			name:       "service validation map into 400",
			method:     http.MethodPost,
			path:       "/todos",
			body:       `{"title": ""}`,
			svc:        &fakeService{err: apperror.ErrInvalidInput},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			New(tt.svc).RegisterRoutes(mux)

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}
