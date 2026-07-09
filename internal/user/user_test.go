package user

import (
	"appointments/internal/assert"
	"appointments/internal/validator"
	"strings"
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
			err := p.Set(tt.plaintext)
			assert.NilError(t, err)

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

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		valid     bool
	}{
		{name: "valid_plaintext", plaintext: "Str0NgP@55word", valid: true},
		{name: "empty_plaintext", plaintext: "", valid: false},
		{name: "max_valid_plaintext", plaintext: strings.Repeat("a", 72), valid: true},
		{name: "less_min_plaintext", plaintext: "abc123", valid: false},
		{name: "greater_max_plaintext", plaintext: strings.Repeat("a", 73), valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidatePassword(v, tt.plaintext)

			assert.Equal(t, v.Valid(), tt.valid)
		})
	}
}

func TestValidateUserAndEmail(t *testing.T) {
	tests := []struct {
		name       string
		user       *User
		wantErrKey string
	}{
		{name: "valid_user", user: &User{
			FirstName:  "New",
			SecondName: "User",
			Email:      "user@test.com",
			Role:       RoleClient,
		}},
		{name: "empty_first_name", user: &User{
			FirstName:  "",
			SecondName: "User",
			Email:      "user@test.com",
			Role:       RoleClient,
		}, wantErrKey: "first_name"},
		{name: "empty_second_name", user: &User{
			FirstName:  "New",
			SecondName: "",
			Email:      "user@test.com",
			Role:       RoleClient,
		}, wantErrKey: "second_name"},
		{name: "greater_max_first_name", user: &User{
			FirstName:  strings.Repeat("a", 65),
			SecondName: "User",
			Email:      "user@test.com",
			Role:       RoleClient,
		}, wantErrKey: "first_name"},
		{name: "greater_max_second_name", user: &User{
			FirstName:  "New",
			SecondName: strings.Repeat("a", 65),
			Email:      "user@test.com",
			Role:       RoleClient,
		}, wantErrKey: "second_name"},
		{name: "max_valid_first_name", user: &User{
			FirstName:  strings.Repeat("a", 64),
			SecondName: "User",
			Email:      "user@test.com",
			Role:       RoleClient,
		}},
		{name: "max_valid_second_name", user: &User{
			FirstName:  "New",
			SecondName: strings.Repeat("a", 64),
			Email:      "user@test.com",
			Role:       RoleClient,
		}},
		{name: "empty_email", user: &User{
			FirstName:  "New",
			SecondName: "User",
			Email:      "",
			Role:       RoleClient,
		}, wantErrKey: "email"},
		{name: "invalid_email", user: &User{
			FirstName:  "New",
			SecondName: "User",
			Email:      "inv@lid_em@il.test",
			Role:       RoleClient,
		}, wantErrKey: "email"},
		{name: "empty_role", user: &User{
			FirstName:  "New",
			SecondName: "User",
			Email:      "user@test.com",
			Role:       "",
		}, wantErrKey: "role"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			err := tt.user.Password.Set("Str0NgP@55word")
			assert.NilError(t, err)

			ValidateUser(v, *tt.user)

			if tt.wantErrKey != "" {
				_, exists := v.Errors[tt.wantErrKey]
				assert.Equal(t, exists, true)
			}
			assert.Equal(t, v.Valid(), tt.wantErrKey == "")
		})
	}
}
