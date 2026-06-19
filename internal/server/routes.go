package server

import (
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", s.healthcheck)
	mux.HandleFunc("POST /v1/users", s.users.Register)
	mux.HandleFunc("PUT /v1/verify", s.users.Verify)

	return mux
}
