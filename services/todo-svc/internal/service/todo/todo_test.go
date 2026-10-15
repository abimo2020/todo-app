package todo

import (
	"errors"
	"testing"

	"github.com/abimo2020/todo-svc/apperror"
	"github.com/abimo2020/todo-svc/internal/model"
)

var (
	errDB     = errors.New("db down")
	errBroker = errors.New("broker down")
)

type fakeRepo struct {
	todos     map[string]model.Todo
	createErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{todos: make(map[string]model.Todo)}
}

func (f *fakeRepo) GetList() ([]model.Todo, error) {
	result := make([]model.Todo, 0, len(f.todos))
	for _, v := range f.todos {
		result = append(result, v)
	}
	return result, nil
}

func (f *fakeRepo) GetDetail(id string) (model.Todo, error) {
	t, ok := f.todos[id]
	if !ok {
		return model.Todo{}, apperror.ErrNotFound
	}
	return t, nil
}

func (f *fakeRepo) Create(t model.Todo) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.todos[t.ID] = t
	return nil
}

func (f *fakeRepo) Update(t model.Todo) error {
	if _, ok := f.todos[t.ID]; !ok {
		return apperror.ErrNotFound
	}
	f.todos[t.ID] = t
	return nil
}

func (f *fakeRepo) Delete(id string) error {
	if _, ok := f.todos[id]; !ok {
		return apperror.ErrNotFound
	}
	delete(f.todos, id)
	return nil
}

type fakeNotifier struct {
	calls int
	err   error
}

func (f *fakeNotifier) NotifyTodoCreated(t model.Todo) error {
	f.calls++
	return f.err
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		repoErr      error
		notifierErr  error
		wantErr      error
		wantSaved    int
		wantNotified int
	}{
		{
			name:    "empty title rejected",
			title:   "",
			wantErr: apperror.ErrEmptyTitle,
		},
		{
			name:         "failed to send notification doesn't prevent data creation",
			title:        "Testing 1",
			wantSaved:    1,
			wantNotified: 1,
			notifierErr:  errBroker,
		}, {
			name:    "failed to create data",
			title:   "Testing 2",
			wantErr: errDB,
			repoErr: errDB,
		},
		{
			name:         "success to create todo",
			title:        "Test ok",
			wantSaved:    1,
			wantNotified: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.createErr = tt.repoErr
			notif := &fakeNotifier{}
			notif.err = tt.notifierErr
			svc := New(repo, notif)

			got, err := svc.Create(model.Todo{Title: tt.title})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if len(repo.todos) != tt.wantSaved {
				t.Fatalf("saved todo = %d, want %d", len(repo.todos), tt.wantSaved)
			}
			if notif.calls != tt.wantNotified {
				t.Fatalf("notify = %d, want %d", notif.calls, tt.wantNotified)
			}
			if tt.wantErr == nil {
				if got.ID == "" {
					t.Error("Required ID")
				}
				if got.CreatedAt.IsZero() {
					t.Error("Required Created At")
				}
			}
		})
	}
}

func TestMarkDone(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		repoErr  error
		wantErr  error
		wantDone bool
	}{
		{
			name:     "Todo doesn't exist",
			id:       "Ngga ada",
			wantErr:  apperror.ErrNotFound,
			wantDone: false,
		},
		{
			name:     "Success to markdone",
			id:       "ada",
			wantDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.createErr = tt.repoErr
			repo.todos["ada"] = model.Todo{ID: "ada", Title: "Test ada"}
			svc := New(repo, &fakeNotifier{})

			err := svc.MarkDone(tt.id)

			if tt.wantDone && !repo.todos[tt.id].Done {
				t.Fatalf("want done %v, todo done %v", tt.wantDone, repo.todos[tt.id].Done)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
