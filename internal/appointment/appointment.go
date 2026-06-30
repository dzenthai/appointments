package appointment

import (
	"appointments/internal/filters"
	"appointments/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	v.Check(len(apt.Title) <= 32, "title", "must not be more than 32 chars long")
	v.Check(len(apt.Description) <= 256, "description", "must not be more than 256 chars long")
	v.Check(apt.ClientID != apt.ProviderID, "client_id", "client ID must not match provider ID")
	v.Check(apt.Title != "", "title", "must be provided")
	v.Check(apt.StartsAt.After(time.Now()), "starts_at", "starts at must be greater that now")
	v.Check(apt.EndsAt.After(apt.StartsAt), "ends_at", "ends at must be greater that starts at")
}

func ValidateStatus(v *validator.Validator, from, to Status) {
	v.Check(canTransition(from, to), "status", "invalid status transition")
}

func canTransition(from, to Status) bool {
	switch from {
	case StatusScheduled:
		return to == StatusConfirmed || to == StatusCancelled || to == StatusScheduled
	case StatusConfirmed:
		return to == StatusCompleted || to == StatusCancelled || to == StatusScheduled
	case StatusCompleted, StatusCancelled:
		return false
	}
	return false
}

func (s *Store) GetAllByClient(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error) {
	query := fmt.Sprintf(
		`SELECT id, client_id, provider_id, title, description, starts_at, ends_at, status, created_at, updated_at, version
       	FROM appointments
       	WHERE client_id = $1 
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`,
		f.SortColumn(), f.SortDirection())

	return s.queryAppointments(ctx, query, userID, f)
}

func (s *Store) GetAllByProvider(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error) {
	query := fmt.Sprintf(
		`SELECT id, client_id, provider_id, title, description, starts_at, ends_at, status, created_at, updated_at, version
       	FROM appointments
       	WHERE provider_id = $1 
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`,
		f.SortColumn(), f.SortDirection())

	return s.queryAppointments(ctx, query, userID, f)
}

func (s *Store) queryAppointments(ctx context.Context, query string, userID int64, f filters.Filters) ([]Appointment, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	args := []any{userID, f.Limit(), f.Offset()}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	apts := make([]Appointment, 0)

	for rows.Next() {
		var apt Appointment
		err = rows.Scan(
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
			return nil, err
		}

		apts = append(apts, apt)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apts, nil
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
		SET provider_id = $1,
		    title = $2,
		    description = $3,
		    starts_at = $4,
		    ends_at = $5,
		    status = $6,
		    updated_at = now(),
		    version = version + 1
		WHERE id = $7 AND version = $8 
		RETURNING provider_id, created_at, updated_at, version`

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	args := []any{apt.ProviderID, apt.Title, apt.Description, apt.StartsAt, apt.EndsAt, apt.Status, apt.ID, apt.Version}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&apt.ProviderID,
		&apt.CreatedAt,
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
