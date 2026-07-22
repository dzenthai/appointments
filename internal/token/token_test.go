package token

import (
	"appointments/internal/assert"
	"appointments/internal/validator"
	"strings"
	"testing"
)

func TestValidateToken(t *testing.T) {

	tests := []struct {
		name       string
		plaintext  string
		wantErrKey string
	}{
		{name: "valid_min_value", plaintext: strings.Repeat("A", 26)},
		{name: "valid_max_value", plaintext: strings.Repeat("A", 52)},
		{name: "empty_plaintext", plaintext: "", wantErrKey: "token"},
		{name: "greater_max_value", plaintext: strings.Repeat("A", 53), wantErrKey: "token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateAuthToken(v, tt.plaintext)
			if tt.wantErrKey != "" {
				_, exists := v.Errors[tt.wantErrKey]
				assert.Equal(t, exists, true)
			}
			assert.Equal(t, v.Valid(), tt.wantErrKey == "")
		})
	}
}
