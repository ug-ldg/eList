package model

import "time"

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	ParentID  *int      `json:"parent_id,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskNode struct {
	ID        int         `json:"id"`
	Title     string      `json:"title"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Children  []*TaskNode `json:"children"`
}
