package server

import (
	"appointments/internal/jsonutil"
	"fmt"
	"net/http"
)

func (s *Server) panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				jsonutil.ServerErrorResponse(w, r, fmt.Errorf("%s", err), s.logger)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
