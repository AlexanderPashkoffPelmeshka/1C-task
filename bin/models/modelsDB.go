package models

import "time"

type Case struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	PromisedAt  time.Time  `json:"promised_at"`
	DueAt       time.Time  `json:"due_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
}

type Comment struct {
	ID        int64     `json:"id"`
	CaseID    int64     `json:"case_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	StatusOpen      = "open"
	StatusDone      = "done"
	StatusCancelled = "cancelled"
)
