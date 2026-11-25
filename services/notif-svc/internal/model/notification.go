package model

import (
	"errors"
	"time"
)

var ErrDuplicate = errors.New("notification already exist")

type Notification struct {
	ID        string    `bson:"_id"`
	Type      string    `bson:"type"`
	TodoID    string    `bson:"todo_id"`
	Title     string    `bson:"title"`
	CreatedAt time.Time `bson:"created_at"`
}
