package store

import (
	"appointments/internal/appointment"
	"appointments/internal/token"
	"appointments/internal/user"
	"database/sql"
)

type Store struct {
	User        *user.Store
	Token       *token.Store
	Appointment *appointment.Store
}

func New(db *sql.DB) *Store {
	return &Store{
		User:        user.NewStore(db),
		Token:       token.NewStore(db),
		Appointment: appointment.NewStore(db),
	}
}
