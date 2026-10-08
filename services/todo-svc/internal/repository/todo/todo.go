package todo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/abimo2020/todo-svc/apperror"
	"github.com/abimo2020/todo-svc/internal/model"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (t *Repository) GetList() ([]model.Todo, error) {
	query := "SELECT id, title, done, created_at FROM todos WHERE deleted_at IS null ORDER BY created_at desc"
	rows, err := t.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		var todo model.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}
func (t *Repository) GetDetail(id string) (model.Todo, error) {
	query := "SELECT id, title, done, created_at FROM todos WHERE id = $1"

	var todo model.Todo
	err := t.db.QueryRowContext(context.Background(), query, id).Scan(
		&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Todo{}, apperror.ErrNotFound
	}
	if err != nil {
		return model.Todo{}, err
	}
	return todo, nil
}
func (t *Repository) Create(todo model.Todo) error {
	query := "INSERT INTO todos (id, title, done, created_at) VALUEs ($1,$2,$3,$4)"
	_, err := t.db.ExecContext(context.Background(), query,
		&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
func (t *Repository) Update(todo model.Todo) error {
	query := "UPDATE todos SET title = $1, done = $2, updated_at = $3 WHERE id = $4"
	res, err := t.db.ExecContext(context.Background(), query,
		todo.Title, todo.Done, todo.UpdateAt, todo.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
func (t *Repository) Delete(id string) error {
	query := "UPDATE todos SET deleted_at = $1 WHERE id = $2"
	res, err := t.db.ExecContext(context.Background(), query, time.Now(), id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
