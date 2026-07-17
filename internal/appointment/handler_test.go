package appointment

import (
	"appointments/internal/assert"
	"appointments/internal/user"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestList(t *testing.T) {

	client := newUser(100, user.RoleClient)

	provider := newUser(200, user.RoleProvider)

	admin := newUser(1, user.RoleAdmin)

	anonymous := user.AnonymousUser

	apts := newApts()

	tests := []struct {
		name       string
		ctxUser    *user.User
		apts       []Appointment
		query      string
		err        error
		wantStatus int
		wantStore  user.Role
	}{
		{name: "client_list", ctxUser: client, apts: apts, wantStatus: http.StatusOK, wantStore: user.RoleClient},
		{name: "provider_list", ctxUser: provider, apts: apts, wantStatus: http.StatusOK, wantStore: user.RoleProvider},
		{name: "admin_list", ctxUser: admin, apts: apts, wantStatus: http.StatusOK, wantStore: user.RoleAdmin},
		{name: "anonymous_list", ctxUser: anonymous, apts: apts, wantStatus: http.StatusNotFound},
		{name: "valid_query", ctxUser: client, apts: apts, query: "?sort=title&page=2&page_size=10", wantStatus: http.StatusOK, wantStore: user.RoleClient},
		{name: "invalid_query", ctxUser: client, query: "?title=id&page=-1&page_size=-20", wantStatus: http.StatusBadRequest, wantStore: ""},
		{name: "store_error", ctxUser: client, err: errors.New("mock error"), wantStatus: http.StatusInternalServerError, wantStore: user.RoleClient},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			store := &storeMock{
				apts:   tt.apts,
				getErr: tt.err,
			}

			h := NewHandler(store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
			req = user.SetUserContext(req, tt.ctxUser)

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
		ctxUser    *user.User
		apt        *Appointment
		param      string
		err        error
		wantStatus int
	}{
		{name: "client_valid", ctxUser: client, apt: clientApt, param: "1", wantStatus: http.StatusOK},
		{name: "provider_valid", ctxUser: provider, apt: clientApt, param: "1", wantStatus: http.StatusOK},
		{name: "admin_valid", ctxUser: admin, apt: foreignApt, param: "2", wantStatus: http.StatusOK},
		{name: "client_invalid_param", ctxUser: client, apt: clientApt, param: "abc", wantStatus: http.StatusBadRequest},
		{name: "not_found", ctxUser: client, apt: clientApt, param: "100", err: ErrAppointmentNotFound, wantStatus: http.StatusNotFound},
		{name: "client_access_denied", ctxUser: client, apt: foreignApt, param: "2", wantStatus: http.StatusNotFound},
		{name: "provider_access_denied", ctxUser: provider, apt: foreignApt, param: "2", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &storeMock{apt: tt.apt, getErr: tt.err}

			h := NewHandler(store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.SetPathValue("id", tt.param)
			req = user.SetUserContext(req, tt.ctxUser)

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
		ID:         userID,
		FirstName:  "first name",
		SecondName: "second name",
		Email:      fmt.Sprintf("%s@test.com", role),
		Role:       role,
		Verified:   true,
		CreatedAt:  time.Time{},
		Version:    1,
	}
}

func newApt(aptID, clientID, providerID int64, status Status) *Appointment {
	now := time.Now()
	return &Appointment{
		ID:         aptID,
		ClientID:   clientID,
		ProviderID: providerID,
		Title:      "title",
		StartsAt:   now.Add(1 * time.Hour),
		EndsAt:     now.Add(2 * time.Hour),
		Status:     status,
		Version:    1,
	}
}

func newApts() []Appointment {
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
