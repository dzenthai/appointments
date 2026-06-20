package store

import (
	"appointments/internal/token"
	"appointments/internal/user"
	"database/sql"
)

type Store struct {
	User         *user.Store
	Verification *token.Store
}

func New(db *sql.DB) *Store {
	return &Store{
		User:         user.NewStore(db),
		Verification: token.NewStore(db),
	}
}
