package server

import (
	"net/http"
)

func (s *Server) route() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", s.healthcheck)

	return mux
}
