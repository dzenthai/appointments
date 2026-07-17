package user

import (
	"appointments/internal/token"
	"context"
)

type storeMock struct {
	user      *User
	getErr    error
	createErr error
	updateErr error
}

func (s *storeMock) GetByToken(ctx context.Context, plaintext string, scope token.Scope) (*User, error) {
	return s.user, s.getErr
}

func (s *storeMock) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.user, s.getErr
}

func (s *storeMock) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.user, s.getErr
}

func (s *storeMock) Insert(ctx context.Context, user *User) error {
	return s.createErr
}

func (s *storeMock) Update(ctx context.Context, user *User) error {
	return s.updateErr
}
