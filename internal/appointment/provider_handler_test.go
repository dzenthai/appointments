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
)

func TestConfirm(t *testing.T) {

	foreignClient := newUser(101, user.RoleClient)

	provider := newUser(200, user.RoleProvider)

	foreignProvider := newUser(201, user.RoleProvider)

	apt := newApt(1, 100, 200, StatusScheduled)

	cancelledApt := newApt(1, 100, 200, StatusCancelled)

	confirmedApt := newApt(2, 100, 200, StatusConfirmed)

	completedApt := newApt(3, 100, 200, StatusCompleted)

	tests := []struct {
		name       string
		ctxUser    *user.User
		apt        *Appointment
		param      string
		getErr     error
		updateErr  error
		wantStatus int
	}{
		{name: "valid_confirmation", ctxUser: provider, apt: apt, param: "1", wantStatus: http.StatusOK},
		{name: "not_found", ctxUser: provider, apt: apt, param: "2", getErr: ErrAppointmentNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid_param", ctxUser: provider, apt: apt, param: "abc", wantStatus: http.StatusBadRequest},
		{name: "foreign_provider_access_denied", ctxUser: foreignProvider, apt: apt, param: "1", wantStatus: http.StatusNotFound},
		{name: "client_access_denied", ctxUser: foreignClient, apt: apt, param: "1", wantStatus: http.StatusNotFound},
		{name: "edit_conflict", ctxUser: provider, apt: apt, param: "1", updateErr: ErrEditConflict, wantStatus: http.StatusConflict},
		{name: "cancel_transition", ctxUser: provider, apt: cancelledApt, param: "1", wantStatus: http.StatusBadRequest},
		{name: "confirmed_transition", ctxUser: provider, apt: confirmedApt, param: "2", wantStatus: http.StatusBadRequest},
		{name: "completed_transition", ctxUser: provider, apt: completedApt, param: "3", wantStatus: http.StatusBadRequest},
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
			req = user.SetUserContext(req, tt.ctxUser)

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
