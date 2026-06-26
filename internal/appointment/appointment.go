package appointment

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrAppointmentNotFound = errors.New("appointment not found")
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

func (s *Store) GetByID(id int64) (*Appointment, error) {
	query :=
		`SELECT client_id, provider_id, title, description, starts_at, ends_at, status, created_at, updated_at, version
		FROM appointments
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var app Appointment

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&app.ClientID,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrAppointmentNotFound
		default:
			return nil, err
		}
	}

	return &app, nil
}

func (s *Store) Insert(app *Appointment) error {
	return nil
}
