package user

import (
	"appointments/internal/token"
	"context"
)

type StoreMock struct {
	user      *User
	getErr    error
	createErr error
	updateErr error
}

func (s *StoreMock) GetByToken(ctx context.Context, plaintext string, scope token.Scope) (*User, error) {
	return s.user, s.getErr
}

func (s *StoreMock) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.user, s.getErr
}

func (s *StoreMock) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.user, s.getErr
}

func (s *StoreMock) Insert(ctx context.Context, user *User) error {
	return s.createErr
}

func (s *StoreMock) Update(ctx context.Context, user *User) error {
	return s.updateErr
}
