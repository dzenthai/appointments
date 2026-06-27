package server

import (
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", s.healthcheck)

	mux.HandleFunc("POST /v1/users", s.users.Register)
	mux.HandleFunc("PUT /v1/users/verify", s.users.Verify)
	mux.HandleFunc("POST /v1/users/login", s.users.Login)

	mux.HandleFunc("GET /v1/appointments/{id}", s.requireAuthentication(s.appHandler.Show))
	mux.HandleFunc("POST /v1/appointments", s.requireVerification(s.appHandler.Create))

	return s.recoverPanic(s.authenticate(mux))
}
