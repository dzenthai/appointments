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
)

type mockStore struct {
	apt       *Appointment
	apts      []Appointment
	role      user.Role
	getErr    error
	updateErr error
}

func (s *mockStore) GetAllByAdmin(ctx context.Context, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleAdmin
	return s.apts, s.getErr
}

func (s *mockStore) GetAllByClient(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleClient
	return s.apts, s.getErr
}

func (s *mockStore) GetAllByProvider(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error) {
	s.role = user.RoleProvider
	return s.apts, s.getErr
}

func (s *mockStore) GetByID(ctx context.Context, id int64) (*Appointment, error) {
	return s.apt, s.getErr
}

func (s *mockStore) Insert(ctx context.Context, apt *Appointment) error {
	return s.getErr
}

func (s *mockStore) Update(ctx context.Context, apt *Appointment) error {
	return s.updateErr
}

func TestList(t *testing.T) {

	client := newUser(100, user.RoleClient)

	provider := newUser(200, user.RoleProvider)

	admin := newUser(1, user.RoleAdmin)

	anonymous := user.AnonymousUser

	apts := getApts()

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
				apts:   tt.apts,
				getErr: tt.err,
			}

			h := NewHandler(store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
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

	client := newUser(100, user.RoleClient)

	provider := newUser(200, user.RoleProvider)

	admin := newUser(1, user.RoleAdmin)

	clientApt := newApt(1, 100, 200, StatusScheduled)

	foreignApt := newApt(2, 101, 202, StatusConfirmed)

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
			store := &mockStore{apt: tt.apt, getErr: tt.err}

			h := NewHandler(store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
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

func newUser(userID int64, role user.Role) *user.User {
	return &user.User{
		ID:   userID,
		Role: role,
	}
}

func newApt(aptID, clientID, providerID int64, status Status) *Appointment {
	return &Appointment{
		ID:         aptID,
		ClientID:   clientID,
		ProviderID: providerID,
		Status:     status,
	}
}

func getApts() []Appointment {
	return []Appointment{
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
}
