package appointment

import (
	"appointments/internal/validator"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrEditConflict        = errors.New("edit conflict")
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

func ValidateAppointment(v *validator.Validator, apt *Appointment) {
	v.Check(apt.ClientID != apt.ProviderID, "client_id", "client ID must not match provider ID")
	v.Check(apt.Title != "", "title", "must be provided")
	v.Check(apt.EndsAt.After(apt.StartsAt), "ends_at", "ends at must be greater that starts at")
}

func (s *Store) GetByID(ctx context.Context, id int64) (*Appointment, error) {
	query :=
		`SELECT id, client_id, provider_id, title, description, starts_at, ends_at, status, created_at, updated_at, version
		FROM appointments
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	var apt Appointment

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&apt.ID,
		&apt.ClientID,
		&apt.ProviderID,
		&apt.Title,
		&apt.Description,
		&apt.StartsAt,
		&apt.EndsAt,
		&apt.Status,
		&apt.CreatedAt,
		&apt.UpdatedAt,
		&apt.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrAppointmentNotFound
		default:
			return nil, err
		}
	}

	return &apt, nil
}

func (s *Store) Insert(ctx context.Context, apt *Appointment) error {
	query :=
		`INSERT INTO appointments (client_id, provider_id, title, description, starts_at, ends_at, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at, version`

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	args := []any{apt.ClientID, apt.ProviderID, apt.Title, apt.Description, apt.StartsAt, apt.EndsAt, apt.Status}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&apt.ID,
		&apt.CreatedAt,
		&apt.UpdatedAt,
		&apt.Version,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Update(ctx context.Context, apt *Appointment) error {
	query :=
		`UPDATE appointments 
		SET title = $1,
		    description = $2,
		    starts_at = $3,
		    ends_at = $4,
		    status = $5,
		    updated_at = now(),
		    version = version + 1
		WHERE id = $6 AND version = $7 
		RETURNING updated_at, version`

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	args := []any{apt.Title, apt.Description, apt.StartsAt, apt.EndsAt, apt.Status, apt.UpdatedAt}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&apt.UpdatedAt,
		&apt.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
