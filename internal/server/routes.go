package server

import (
	"appointments/internal/user"
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", s.healthcheck)

	mux.HandleFunc("POST /v1/users", s.users.Register)
	mux.HandleFunc("PUT /v1/users/verify", s.users.Verify)
	mux.HandleFunc("POST /v1/users/login", s.users.Login)

	mux.HandleFunc("GET /v1/appointments/{id}", s.requireAuthentication(s.appHandler.Show))

	mux.HandleFunc("GET /v1/appointments", s.requireVerification(s.appHandler.List))

	mux.HandleFunc("POST /v1/appointments", s.requireRole(user.RoleClient, s.appHandler.Create))
	mux.HandleFunc("PATCH /v1/appointments/{id}", s.requireRole(user.RoleClient, s.appHandler.Update))
	mux.HandleFunc("PATCH /v1/appointments/{id}/cancel", s.requireRole(user.RoleClient, s.appHandler.Cancel))

	mux.HandleFunc("PATCH /v1/appointments/{id}/confirm", s.requireRole(user.RoleProvider, s.appHandler.Confirm))

	return s.recoverPanic(s.enableCORS(s.rateLimiter(s.authenticate(mux))))
}
