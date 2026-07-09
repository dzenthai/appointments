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

func TestPasswordSetAndMatches(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{name: "plaintext_password", plaintext: "pa55word"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := new(password)
			if err := p.Set(tt.plaintext); err != nil {
				t.Fatalf("error while generating hash of password: %s", err.Error())
			}
			match, err := p.Matches(tt.plaintext)
			assert.NilError(t, err)
			assert.Equal(t, match, true)

			match, err = p.Matches("wr0ng_pa55word")
			assert.NilError(t, err)
			assert.Equal(t, match, false)

			assert.Equal(t, len(p.hash) > 0, true)
			assert.Equal(t, *p.plaintext, tt.plaintext)
		})
	}
}
