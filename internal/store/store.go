package store

import (
	"appointments/internal/user"
	"database/sql"
)

type Store struct {
	User *user.Store
}

func New(db *sql.DB) *Store {
	return &Store{
		User: user.NewStore(db),
	}
}
