package server

import (
	"appointments/internal/jsonutil"
	"appointments/internal/vcs"
	"net/http"
)

type Healthcheck struct {
	Status  string `json:"status"`
	Env     string `json:"env"`
	Version string `json:"version"`
}

func (s *Server) healthcheck(w http.ResponseWriter, r *http.Request) {
	hc := Healthcheck{
		Status:  "OK",
		Env:     s.cfg.Env,
		Version: vcs.Version(),
	}
	err := jsonutil.WriteJSON(w, http.StatusOK, hc, nil)
	if err != nil {
		return
	}
}
