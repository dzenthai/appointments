package user

import (
	"appointments/internal/assert"
	"testing"
)

func TestIsAnonymous(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want bool
	}{
		{name: "anonymous_user", user: AnonymousUser, want: true},
		{name: "existed_user", user: &User{ID: 1}, want: false},
		{name: "zero_value_user", user: new(User), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.user.IsAnonymous(), tt.want)
		})
	}
}
