package appointment

import (
	"appointments/internal/assert"
	"appointments/internal/user"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConfirm(t *testing.T) {

	client := &user.User{
		ID:         101,
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

	foreignProvider := &user.User{
		ID:         201,
		FirstName:  "Foreign",
		SecondName: "Provider",
		Email:      "foreign@test.com",
		Role:       user.RoleProvider,
		Verified:   true,
		CreatedAt:  time.Now(),
		Version:    1,
	}

	apt := &Appointment{
		ID:          1,
		ClientID:    100,
		ProviderID:  200,
		Title:       "first appointment",
		Description: "description",
		Status:      StatusScheduled,
		Version:     1,
	}

	cancelledApt := &Appointment{
		ID:          1,
		ClientID:    100,
		ProviderID:  200,
		Title:       "cancelled",
		Description: "description",
		Status:      StatusCancelled,
		Version:     1,
	}

	confirmedApt := &Appointment{
		ID:          1,
		ClientID:    100,
		ProviderID:  200,
		Title:       "confirmed",
		Description: "description",
		Status:      StatusConfirmed,
		Version:     1,
	}

	completedApt := &Appointment{
		ID:          1,
		ClientID:    100,
		ProviderID:  200,
		Title:       "completed",
		Description: "description",
		Status:      StatusCompleted,
		Version:     1,
	}

	tests := []struct {
		name       string
		user       *user.User
		apt        *Appointment
		param      string
		getErr     error
		updateErr  error
		wantStatus int
	}{
		{name: "valid_confirmation", user: provider, apt: apt, param: "1", wantStatus: http.StatusOK},
		{name: "not_found", user: provider, apt: apt, param: "2", getErr: ErrAppointmentNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid_param", user: provider, apt: apt, param: "abc", wantStatus: http.StatusBadRequest},
		{name: "foreign_provider_access_denied", user: foreignProvider, apt: apt, param: "1", wantStatus: http.StatusNotFound},
		{name: "client_access_denied", user: client, apt: apt, param: "1", wantStatus: http.StatusNotFound},
		{name: "edit_conflict", user: provider, apt: apt, param: "1", updateErr: ErrEditConflict, wantStatus: http.StatusConflict},
		{name: "cancel_transition", user: provider, apt: cancelledApt, param: "1", wantStatus: http.StatusBadRequest},
		{name: "confirmed_transition", user: provider, apt: confirmedApt, param: "1", wantStatus: http.StatusBadRequest},
		{name: "completed_transition", user: provider, apt: completedApt, param: "1", wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{
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
