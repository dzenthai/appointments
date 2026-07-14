package appointment

import (
	"appointments/internal/assert"
	"appointments/internal/filters"
	"appointments/internal/user"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockStore struct {
	apt  *Appointment
	apts []Appointment
	role user.Role
	err  error
}

func (s *mockStore) GetAllByAdmin(ctx context.Context, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleAdmin
	return s.apts, s.err
}

func (s *mockStore) GetAllByClient(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleClient
	return s.apts, s.err
}

func (s *mockStore) GetAllByProvider(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleProvider
	return s.apts, s.err
}

func (s *mockStore) GetByID(ctx context.Context, id int64) (*Appointment, error) {
	return s.apt, s.err
}

func (s *mockStore) Insert(ctx context.Context, apt *Appointment) error {
	return s.err
}

func (s *mockStore) Update(ctx context.Context, apt *Appointment) error {
	return s.err
}

func TestList(t *testing.T) {

	client := &user.User{
		ID:         100,
		FirstName:  "New",
		SecondName: "Client",
		Email:      "client@test.com",
		Role:       user.RoleClient,
		Verified:   true,
		CreatedAt:  time.Now(),
		Version:    1,
	}

	provider := &user.User{
		ID:         200,
		FirstName:  "New",
		SecondName: "Provider",
		Email:      "provider@test.com",
		Role:       user.RoleProvider,
		Verified:   true,
		CreatedAt:  time.Now(),
		Version:    1,
	}

	admin := &user.User{
		ID:         1,
		FirstName:  "New",
		SecondName: "Admin",
		Email:      "admin@test.com",
		Role:       user.RoleAdmin,
		Verified:   true,
		CreatedAt:  time.Now(),
		Version:    1,
	}

	anonymous := user.AnonymousUser

	apts := []Appointment{
		{
			ClientID:   100,
			ProviderID: 200,
		},
		{
			ClientID:   100,
			ProviderID: 201,
		},
		{
			ClientID:   100,
			ProviderID: 202,
		},
		{
			ClientID:   101,
			ProviderID: 200,
		},
		{
			ClientID:   102,
			ProviderID: 200,
		},
		{
			ClientID:   103,
			ProviderID: 200,
		},
	}

	tests := []struct {
		name       string
		user       *user.User
		apts       []Appointment
		query      string
		err        error
		wantStatus int
		wantStore  user.Role
	}{
		{name: "client_list", user: client, apts: apts, wantStatus: http.StatusOK, wantStore: user.RoleClient},
		{name: "provider_list", user: provider, apts: apts, wantStatus: http.StatusOK, wantStore: user.RoleProvider},
		{name: "admin_list", user: admin, apts: apts, wantStatus: http.StatusOK, wantStore: user.RoleAdmin},
		{name: "anonymous_list", user: anonymous, apts: apts, wantStatus: http.StatusNotFound},
		{name: "valid_query", user: client, apts: apts, query: "?sort=title&page=2&page_size=10", wantStatus: http.StatusOK, wantStore: user.RoleClient},
		{name: "invalid_query", user: client, query: "?title=id&page=-1&page_size=-20", wantStatus: http.StatusBadRequest, wantStore: ""},
		{name: "store_error", user: client, err: errors.New("mock error"), wantStatus: http.StatusInternalServerError, wantStore: user.RoleClient},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			store := &mockStore{
				apts: tt.apts,
				err:  tt.err,
			}

			h := NewHandler(store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/v1/appointments"+tt.query, nil)
			req = user.SetUserContext(req, tt.user)

			h.List(rec, req)

			assert.Equal(t, rec.Code, tt.wantStatus)
			assert.Equal(t, store.role, tt.wantStore)

			if tt.wantStatus == http.StatusOK {
				actual, err := io.ReadAll(rec.Body)
				assert.NilError(t, err)
				actual = bytes.TrimSpace(actual)

				expected, err := json.Marshal(tt.apts)
				assert.NilError(t, err)
				assert.Equal(t, string(actual), string(expected))
			}
		})
	}
}

func TestShow(t *testing.T) {

	client := &user.User{
		ID:         100,
		FirstName:  "New",
		SecondName: "User",
		Email:      "user@test.com",
		Role:       user.RoleClient,
		Verified:   true,
		CreatedAt:  time.Now(),
		Version:    1,
	}

	provider := &user.User{
		ID:         200,
		FirstName:  "New",
		SecondName: "Provider",
		Email:      "provider@test.com",
		Role:       user.RoleProvider,
		Verified:   true,
		CreatedAt:  time.Now(),
		Version:    1,
	}

	admin := &user.User{
		ID:         1,
		FirstName:  "New",
		SecondName: "Admin",
		Email:      "admin@test.com",
		Role:       user.RoleAdmin,
		Verified:   true,
		CreatedAt:  time.Now(),
		Version:    1,
	}

	clientApt := &Appointment{
		ID:          1,
		ClientID:    100,
		ProviderID:  200,
		Title:       "first appointment",
		Description: "description",
		Status:      StatusScheduled,
		Version:     1,
	}

	foreignApt := &Appointment{
		ID:          2,
		ClientID:    101,
		ProviderID:  202,
		Title:       "second appointment",
		Description: "description",
		Status:      StatusConfirmed,
		Version:     1,
	}

	tests := []struct {
		name       string
		user       *user.User
		apt        *Appointment
		param      string
		err        error
		wantStatus int
	}{
		{name: "client_valid", user: client, apt: clientApt, param: "1", wantStatus: http.StatusOK},
		{name: "provider_valid", user: provider, apt: clientApt, param: "1", wantStatus: http.StatusOK},
		{name: "admin_valid", user: admin, apt: foreignApt, param: "2", wantStatus: http.StatusOK},
		{name: "client_invalid_param", user: client, apt: clientApt, param: "abc", wantStatus: http.StatusBadRequest},
		{name: "not_found", user: client, apt: clientApt, param: "100", err: ErrAppointmentNotFound, wantStatus: http.StatusNotFound},
		{name: "client_access_denied", user: client, apt: foreignApt, param: "2", wantStatus: http.StatusNotFound},
		{name: "provider_access_denied", user: provider, apt: foreignApt, param: "2", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{apt: tt.apt, err: tt.err}

			h := NewHandler(store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/v1/appointments", nil)
			req.SetPathValue("id", tt.param)
			req = user.SetUserContext(req, tt.user)

			h.Show(rec, req)

			assert.Equal(t, rec.Code, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				actual, err := io.ReadAll(rec.Body)
				assert.NilError(t, err)
				actual = bytes.TrimSpace(actual)

				expected, err := json.Marshal(&store.apt)
				assert.NilError(t, err)
				assert.Equal(t, string(actual), string(expected))
			}
		})
	}
}
