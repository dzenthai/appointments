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
	"strings"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	client := newUser(100, user.RoleClient)

	provider := newUser(200, user.RoleProvider)

	wrongProvider := newUser(201, user.RoleClient)

	admin := newUser(1, user.RoleAdmin)

	validBody := fmt.Sprintf(`{
    	"provider_id": 200,
    	"title": "title",
    	"description": "description",
    	"starts_at": %q,
    	"ends_at": %q
	}`,
		time.Now().Add(time.Hour).Format(time.RFC3339),
		time.Now().Add(2*time.Hour).Format(time.RFC3339),
	)

	invalidBody := fmt.Sprintf(`{
    	"provider_id": 200,
    	"title": "title",
    	"description": "description",
    	"starts_at": %q,
    	"ends_at": %q
	}`,
		time.Date(1999, time.January, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		time.Date(1990, time.January, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
	)

	malformedBody := `{
    	"provider_id": 200,
    	"title": "Example title",
    	"description": "Example description",
    	"starts_at": %q,
    	"ends_at`

	tests := []struct {
		name         string
		ctxUser      *user.User
		storeUser    *user.User
		body         string
		getUserErr   error
		createAptErr error
		wantStatus   int
	}{
		{name: "valid_json_body", ctxUser: client, storeUser: provider, body: validBody, wantStatus: http.StatusCreated},
		{name: "invalid_json_body", ctxUser: client, storeUser: provider, body: invalidBody, wantStatus: http.StatusBadRequest},
		{name: "malformed_json_body", ctxUser: client, storeUser: provider, body: malformedBody, wantStatus: http.StatusBadRequest},
		{name: "non_client_creates", ctxUser: admin, storeUser: provider, body: validBody, wantStatus: http.StatusBadRequest},
		{name: "provider_not_found", ctxUser: client, body: validBody, getUserErr: user.ErrUserNotFound, wantStatus: http.StatusBadRequest},
		{name: "provider_wrong_role", ctxUser: client, storeUser: wrongProvider, body: validBody, wantStatus: http.StatusBadRequest},
		{name: "insert_error", ctxUser: client, storeUser: provider, body: validBody, createAptErr: errors.New("server error"),
			wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aptStore := &storeMock{
				createErr: tt.createAptErr,
			}

			userStore := &userStoreMock{
				user:   tt.storeUser,
				getErr: tt.getUserErr,
			}

			h := NewHandler(aptStore, userStore, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req = user.SetUserContext(req, tt.ctxUser)

			h.Create(rec, req)

			assert.Equal(t, rec.Code, tt.wantStatus)

			if tt.wantStatus == http.StatusCreated {
				var got Appointment
				assert.NilError(t, json.NewDecoder(rec.Body).Decode(&got))

				assert.Equal(t, got.ClientID, tt.ctxUser.ID)
				assert.Equal(t, got.ProviderID, tt.storeUser.ID)
				assert.Equal(t, got.Status, StatusScheduled)
			}
		})
	}
}

func TestCancel(t *testing.T) {

	client := newUser(100, user.RoleClient)

	foreignClient := newUser(101, user.RoleClient)

	foreignProvider := newUser(201, user.RoleProvider)

	apt := newApt(1, 100, 200, StatusScheduled)

	cancelledApt := newApt(1, 100, 200, StatusCancelled)

	confirmedApt := newApt(2, 100, 200, StatusConfirmed)

	completedApt := newApt(3, 100, 200, StatusCompleted)

	tests := []struct {
		name       string
		user       *user.User
		apt        *Appointment
		param      string
		getErr     error
		updateErr  error
		wantStatus int
	}{
		{name: "valid_cancellation", user: client, apt: apt, param: "1", wantStatus: http.StatusOK},
		{name: "not_found", user: client, apt: apt, param: "2", getErr: ErrAppointmentNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid_param", user: client, apt: apt, param: "abc", wantStatus: http.StatusBadRequest},
		{name: "foreign_provider_access_denied", user: foreignProvider, apt: apt, param: "1", wantStatus: http.StatusNotFound},
		{name: "client_access_denied", user: foreignClient, apt: apt, param: "1", wantStatus: http.StatusNotFound},
		{name: "edit_conflict", user: client, apt: apt, param: "1", updateErr: ErrEditConflict, wantStatus: http.StatusConflict},
		{name: "cancel_transition", user: client, apt: cancelledApt, param: "1", wantStatus: http.StatusBadRequest},
		{name: "confirmed_transition", user: client, apt: confirmedApt, param: "2", wantStatus: http.StatusBadRequest},
		{name: "completed_transition", user: client, apt: completedApt, param: "3", wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &storeMock{
				apt:       new(*tt.apt),
				getErr:    tt.getErr,
				updateErr: tt.updateErr,
			}

			h := NewHandler(store, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPatch, "/", nil)
			req.SetPathValue("id", tt.param)
			req = user.SetUserContext(req, tt.user)

			h.Confirm(rec, req)

			assert.Equal(t, rec.Code, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				actual, err := io.ReadAll(rec.Body)
				assert.NilError(t, err)
				actual = bytes.TrimSpace(actual)

				want := *tt.apt
				want.Status = StatusConfirmed

				expected, err := json.Marshal(&want)
				assert.NilError(t, err)

				assert.Equal(t, string(actual), string(expected))
			}
		})
	}
}
