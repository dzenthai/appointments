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
		{name: "provider_access_denied", ctxUser: admin, storeUser: provider, body: validBody, wantStatus: http.StatusBadRequest},
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

			uStore := &userStoreMock{
				user:   tt.storeUser,
				getErr: tt.getUserErr,
			}

			h := NewHandler(aptStore, uStore, slog.New(slog.NewTextHandler(io.Discard, nil)))

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

func TestUpdate(t *testing.T) {

	client := newUser(100, user.RoleClient)

	foreignClient := newUser(101, user.RoleClient)

	provider := newUser(200, user.RoleProvider)

	scheduledApt := newApt(1, client.ID, provider.ID, StatusScheduled)

	confirmedApt := newApt(1, client.ID, provider.ID, StatusConfirmed)

	cancelledApt := newApt(1, client.ID, provider.ID, StatusCancelled)

	completedApt := newApt(1, client.ID, provider.ID, StatusCompleted)

	validBody := fmt.Sprintf(`{
    	"title": "updated title",
    	"description": "updated description",
    	"starts_at": %q,
    	"ends_at": %q
	}`,
		scheduledApt.StartsAt.Add(1*time.Hour).Format(time.RFC3339),
		scheduledApt.EndsAt.Add(2*time.Hour).Format(time.RFC3339),
	)

	titleUpdate := `{
    	"title": "updated title"
	}`

	descriptionUpdate := `{
    	"description": "updated description"
	}`

	startsAtUpdate := fmt.Sprintf(`{
    	"starts_at": %q
	}`,
		scheduledApt.StartsAt.Add(1*time.Hour).Format(time.RFC3339),
	)

	endsAtUpdate := fmt.Sprintf(`{
    	"ends_at": %q
	}`,
		scheduledApt.EndsAt.Add(2*time.Hour).Format(time.RFC3339),
	)

	invalidBody := fmt.Sprintf(`{
    	"title": "updated title",
    	"description": "updated description",
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
		apt          *Appointment
		body         string
		param        string
		getUserErr   error
		getAptErr    error
		updateAptErr error
		wantStatus   int
		check        func(t *testing.T, got, orig Appointment)
	}{
		{name: "valid_json_body", ctxUser: client, apt: scheduledApt, body: validBody, param: "1", wantStatus: http.StatusOK,
			check: func(t *testing.T, got, orig Appointment) {
				t.Helper()
				assert.Equal(t, got.Title, "updated title")
				assert.Equal(t, got.Description, "updated description")
				assert.Equal(t, got.StartsAt.Format(time.RFC3339), orig.StartsAt.Add(1*time.Hour).Format(time.RFC3339))
				assert.Equal(t, got.EndsAt.Format(time.RFC3339), orig.EndsAt.Add(2*time.Hour).Format(time.RFC3339))
			},
		},
		{name: "invalid_json_body", ctxUser: client, apt: scheduledApt, body: invalidBody, param: "1", wantStatus: http.StatusBadRequest},
		{name: "malformed_body", ctxUser: client, apt: scheduledApt, body: malformedBody, param: "1", wantStatus: http.StatusBadRequest},
		{name: "invalid_param", ctxUser: client, apt: scheduledApt, body: validBody, param: "abc", wantStatus: http.StatusBadRequest},
		{name: "not_found", ctxUser: client, apt: scheduledApt, body: validBody, param: "2", getAptErr: ErrAppointmentNotFound, wantStatus: http.StatusNotFound},
		{name: "client_access_denied", ctxUser: foreignClient, apt: scheduledApt, body: malformedBody, param: "1", wantStatus: http.StatusNotFound},
		{name: "edit_conflict", ctxUser: client, apt: scheduledApt, body: validBody, param: "1", updateAptErr: ErrEditConflict, wantStatus: http.StatusConflict},
		{name: "transition_confirmed", ctxUser: client, apt: confirmedApt, body: validBody, param: "1", wantStatus: http.StatusOK},
		{name: "transition_cancelled", ctxUser: client, apt: cancelledApt, body: validBody, param: "1", wantStatus: http.StatusBadRequest},
		{name: "transition_completed", ctxUser: client, apt: completedApt, body: validBody, param: "1", wantStatus: http.StatusBadRequest},
		{name: "title", ctxUser: client, apt: scheduledApt, body: titleUpdate, param: "1", wantStatus: http.StatusOK,
			check: func(t *testing.T, got, orig Appointment) {
				t.Helper()
				assert.Equal(t, got.Title, "updated title")
				assert.Equal(t, got.Description, orig.Description)
				assert.Equal(t, got.StartsAt.Format(time.RFC3339), orig.StartsAt.Format(time.RFC3339))
				assert.Equal(t, got.EndsAt.Format(time.RFC3339), orig.EndsAt.Format(time.RFC3339))
			},
		},
		{name: "description", ctxUser: client, apt: scheduledApt, body: descriptionUpdate, param: "1", wantStatus: http.StatusOK,
			check: func(t *testing.T, got, orig Appointment) {
				t.Helper()
				assert.Equal(t, got.Title, orig.Title)
				assert.Equal(t, got.Description, "updated description")
				assert.Equal(t, got.StartsAt.Format(time.RFC3339), orig.StartsAt.Format(time.RFC3339))
				assert.Equal(t, got.EndsAt.Format(time.RFC3339), orig.EndsAt.Format(time.RFC3339))
			},
		},
		{name: "starts_at", ctxUser: client, apt: scheduledApt, body: startsAtUpdate, param: "1", wantStatus: http.StatusOK,
			check: func(t *testing.T, got, orig Appointment) {
				t.Helper()
				assert.Equal(t, got.Title, orig.Title)
				assert.Equal(t, got.Description, orig.Description)
				assert.Equal(t, got.StartsAt.Format(time.RFC3339), orig.StartsAt.Add(1*time.Hour).Format(time.RFC3339))
				assert.Equal(t, got.EndsAt.Format(time.RFC3339), orig.EndsAt.Format(time.RFC3339))
			},
		},
		{name: "ends_at", ctxUser: client, apt: scheduledApt, body: endsAtUpdate, param: "1", wantStatus: http.StatusOK,
			check: func(t *testing.T, got, orig Appointment) {
				t.Helper()
				assert.Equal(t, got.Title, orig.Title)
				assert.Equal(t, got.Description, orig.Description)
				assert.Equal(t, got.StartsAt.Format(time.RFC3339), orig.StartsAt.Format(time.RFC3339))
				assert.Equal(t, got.EndsAt.Format(time.RFC3339), orig.EndsAt.Add(2*time.Hour).Format(time.RFC3339))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aptStore := &storeMock{
				apt:       new(*tt.apt),
				getErr:    tt.getAptErr,
				updateErr: tt.updateAptErr,
			}

			uStore := &userStoreMock{
				getErr: tt.getUserErr,
			}

			h := NewHandler(aptStore, uStore, slog.New(slog.NewTextHandler(io.Discard, nil)))

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(tt.body))
			req.SetPathValue("id", tt.param)
			req = user.SetUserContext(req, tt.ctxUser)

			h.Update(rec, req)

			assert.Equal(t, rec.Code, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				var got Appointment
				assert.NilError(t, json.NewDecoder(rec.Body).Decode(&got))

				assert.Equal(t, got.ClientID, tt.ctxUser.ID)
				assert.Equal(t, got.Status, StatusScheduled)
				assert.Equal(t, got.Version, 1)

				if tt.check != nil {
					tt.check(t, got, *tt.apt)
				}
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
