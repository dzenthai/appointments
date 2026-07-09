package appointment

import (
	"appointments/internal/assert"
	"appointments/internal/validator"
	"strings"
	"testing"
	"time"
)

func TestValidateAppointment(t *testing.T) {
	tests := []struct {
		name        string
		appointment *Appointment
		wantErrKey  string
	}{
		{name: "valid_appointment", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  2,
			Title:       "title",
			Description: "description",
			StartsAt:    time.Now().Add(time.Hour * 24),
			EndsAt:      time.Now().Add(time.Hour * 25),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     1,
		}},
		{name: "max_valid_title", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  2,
			Title:       strings.Repeat("a", 32),
			Description: "description",
			StartsAt:    time.Now().Add(time.Hour * 24),
			EndsAt:      time.Now().Add(time.Hour * 25),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     2,
		}},
		{name: "greater_max_title", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  2,
			Title:       strings.Repeat("a", 33),
			Description: "description",
			StartsAt:    time.Now().Add(time.Hour * 24),
			EndsAt:      time.Now().Add(time.Hour * 25),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     2,
		}, wantErrKey: "title"},
		{name: "empty_title", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  2,
			Title:       "",
			Description: "description",
			StartsAt:    time.Now().Add(time.Hour * 24),
			EndsAt:      time.Now().Add(time.Hour * 25),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     2,
		}, wantErrKey: "title"},
		{name: "max_valid_description", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  2,
			Title:       "title",
			Description: strings.Repeat("a", 256),
			StartsAt:    time.Now().Add(time.Hour * 24),
			EndsAt:      time.Now().Add(time.Hour * 25),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     2,
		}},
		{name: "greater_max_description", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  2,
			Title:       "title",
			Description: strings.Repeat("a", 257),
			StartsAt:    time.Now().Add(time.Hour * 24),
			EndsAt:      time.Now().Add(time.Hour * 25),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     2,
		}, wantErrKey: "description"},
		{name: "client_id_matches_provider_id", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  1,
			Title:       "title",
			Description: "description",
			StartsAt:    time.Now().Add(time.Hour * 24),
			EndsAt:      time.Now().Add(time.Hour * 25),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     2,
		}, wantErrKey: "client_id"},
		{name: "starts_at_less_or_equals_now", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  2,
			Title:       "title",
			Description: "description",
			StartsAt:    time.Now(),
			EndsAt:      time.Now().Add(time.Hour * 25),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     2,
		}, wantErrKey: "starts_at"},
		{name: "ends_at_less_or_equals_starts_at", appointment: &Appointment{
			ClientID:    1,
			ProviderID:  2,
			Title:       "title",
			Description: "description",
			StartsAt:    time.Now().Add(time.Hour * 24),
			EndsAt:      time.Now().Add(time.Hour * 23),
			Status:      StatusScheduled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Version:     2,
		}, wantErrKey: "ends_at"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateAppointment(v, tt.appointment)

			if tt.wantErrKey != "" {
				_, exists := v.Errors[tt.wantErrKey]
				assert.Equal(t, exists, true)
			}
			assert.Equal(t, v.Valid(), tt.wantErrKey == "")
		})
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "scheduled_to_scheduled", from: StatusScheduled, to: StatusScheduled, want: true},
		{name: "scheduled_to_confirmed", from: StatusScheduled, to: StatusConfirmed, want: true},
		{name: "scheduled_to_cancelled", from: StatusScheduled, to: StatusCancelled, want: true},
		{name: "scheduled_to_completed", from: StatusScheduled, to: StatusCompleted, want: false},

		{name: "confirmed_to_scheduled", from: StatusConfirmed, to: StatusScheduled, want: true},
		{name: "confirmed_to_confirmed", from: StatusConfirmed, to: StatusConfirmed, want: false},
		{name: "confirmed_to_cancelled", from: StatusConfirmed, to: StatusCancelled, want: true},
		{name: "confirmed_to_completed", from: StatusConfirmed, to: StatusCompleted, want: true},

		{name: "cancelled_to_scheduled", from: StatusCancelled, to: StatusScheduled, want: false},
		{name: "cancelled_to_confirmed", from: StatusCancelled, to: StatusConfirmed, want: false},
		{name: "cancelled_to_cancelled", from: StatusCancelled, to: StatusCancelled, want: false},
		{name: "cancelled_to_completed", from: StatusCancelled, to: StatusCompleted, want: false},

		{name: "completed_to_scheduled", from: StatusCompleted, to: StatusScheduled, want: false},
		{name: "completed_to_confirmed", from: StatusCompleted, to: StatusConfirmed, want: false},
		{name: "completed_to_cancelled", from: StatusCompleted, to: StatusCancelled, want: false},
		{name: "completed_to_completed", from: StatusCompleted, to: StatusCompleted, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, canTransition(tt.from, tt.to), tt.want)
		})
	}
}
