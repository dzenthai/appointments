package appointment

import (
	"database/sql"
	"time"
)

type Status string

const (
	StatusScheduled = Status("scheduled")
	StatusConfirmed = Status("confirmed")
	StatusCancelled = Status("cancelled")
	StatusCompleted = Status("completed")
)

type Appointment struct {
	ID          int64     `json:"id"`
	ClientID    int64     `json:"client_id"`
	ProviderID  int64     `json:"provider_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     int       `json:"version"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}
