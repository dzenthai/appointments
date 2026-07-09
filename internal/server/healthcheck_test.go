package server

import (
	"appointments/internal/assert"
	"appointments/internal/config"
	"appointments/internal/vcs"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	cfg := config.Load()
	s := Server{cfg: cfg}

	rec := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
	assert.NilError(t, err)

	s.healthcheck(rec, req)

	res := rec.Result()

	var got healthcheck
	defer res.Body.Close()

	assert.NilError(t, err)

	err = json.NewDecoder(res.Body).Decode(&got)
	assert.NilError(t, err)

	assert.Equal(t, res.StatusCode, http.StatusOK)

	want := healthcheck{
		Status:  "OK",
		Env:     s.cfg.Env,
		Version: vcs.Version(),
	}

	assert.Equal(t, got, want)
}
