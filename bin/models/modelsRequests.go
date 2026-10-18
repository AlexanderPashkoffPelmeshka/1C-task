package models

import "time"

const Localhost = "localhost:8080"

type CaseRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	PromisedAt  time.Time `json:"promised_at"`
	DueAt       time.Time `json:"due_at"`
}

type CloseCaseRequest struct {
	Status string `json:"status"`
}

type CommentRequest struct {
	Body string `json:"body"`
}
