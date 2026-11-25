package todocreated

import (
	"context"
	"errors"
	"testing"

	"github.com/abimo2020/notif-svc/internal/model"
)

type fakeStore struct {
	saved []model.Notification
	err   error
}

func (f *fakeStore) Save(ctx context.Context, notif model.Notification) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, notif)
	return nil
}

func TestProcess(t *testing.T) {
	errDB := errors.New("db down")

	tests := []struct {
		name      string
		storeErr  error
		wantErr   error
		wantSaved int
	}{
		{name: "success to save", wantSaved: 1},
		{name: "duplicate considered successful", storeErr: model.ErrDuplicate},
		{name: "other failures are passed then can be retry", storeErr: errDB, wantErr: errDB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{err: tt.storeErr}
			consumer := New("", store)

			err := consumer.process(context.Background(), model.TodoCreatedEvent{ID: "1", Title: "Belajar"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want = %v", err, tt.wantErr)
			}
			if len(store.saved) != tt.wantSaved {
				t.Errorf("saved = %d, want = %d", len(store.saved), tt.wantSaved)
			}
		})
	}
}
