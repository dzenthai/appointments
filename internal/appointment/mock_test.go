package appointment

import (
	"appointments/internal/filters"
	"appointments/internal/token"
	"appointments/internal/user"
	"context"
)

type storeMock struct {
	apt       *Appointment
	apts      []Appointment
	role      user.Role
	getErr    error
	createErr error
	updateErr error
}

func (s *storeMock) GetAllByAdmin(ctx context.Context, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleAdmin
	return s.apts, s.getErr
}

func (s *storeMock) GetAllByClient(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleClient
	return s.apts, s.getErr
}

func (s *storeMock) GetAllByProvider(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleProvider
	return s.apts, s.getErr
}

func (s *storeMock) GetByID(ctx context.Context, id int64) (*Appointment, error) {
	return s.apt, s.getErr
}

func (s *storeMock) Insert(ctx context.Context, apt *Appointment) error {
	apt.ID = 1
	apt.Version = 1
	return s.createErr
}

func (s *storeMock) Update(ctx context.Context, apt *Appointment) error {
	return s.updateErr
}

type userStoreMock struct {
	user      *user.User
	getErr    error
	createErr error
	updateErr error
}

func (s *userStoreMock) GetByToken(ctx context.Context, plaintext string, scope token.Scope) (*user.User, error) {
	return s.user, s.getErr
}

func (s *userStoreMock) GetByID(ctx context.Context, id int64) (*user.User, error) {
	return s.user, s.getErr
}

func (s *userStoreMock) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.user, s.getErr
}

func (s *userStoreMock) Insert(ctx context.Context, user *user.User) error {
	return s.createErr
}

func (s *userStoreMock) Update(ctx context.Context, user *user.User) error {
	return s.updateErr
}
