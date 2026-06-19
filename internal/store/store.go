package store

import (
	"appointments/internal/user"
	"appointments/internal/verification"
	"database/sql"
)

type Store struct {
	User         *user.Store
	Verification *verification.Store
}

func New(db *sql.DB) *Store {
	return &Store{
		User:         user.NewStore(db),
		Verification: verification.NewStore(db),
	}
}
