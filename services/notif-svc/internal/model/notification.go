package model

type TodoCreatedEvent struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
