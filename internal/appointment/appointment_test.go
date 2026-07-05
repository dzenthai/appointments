package appointment

import (
	"appointments/internal/assert"
	"testing"
)

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
