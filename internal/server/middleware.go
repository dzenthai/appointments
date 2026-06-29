package server

import (
	"appointments/internal/jsonutil"
	"appointments/internal/token"
	"appointments/internal/user"
	"appointments/internal/validator"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for _, trustedOrigin := range s.cfg.CORS.Origins {
				if origin == trustedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions &&
						r.Header.Get("Access-Control-Request-Method") != "" {

						w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusNoContent)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				jsonutil.ServerErrorResponse(w, r, fmt.Errorf("%v", err), s.logger)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		header := r.Header.Get("Authorization")

		if header == "" {
			r = user.SetUserContext(r, user.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			jsonutil.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		authToken := headerParts[1]

		v := validator.New()

		if token.ValidateAuthToken(v, authToken); !v.Valid() {
			jsonutil.FailedValidationResponse(w, v.Errors)
			return
		}

		u, err := s.userStore.GetByToken(r.Context(), authToken, token.ScopeAuthentication)
		if err != nil {
			switch {
			case errors.Is(err, user.ErrUserNotFound):
				jsonutil.InvalidAuthenticationTokenResponse(w, r)
			default:
				jsonutil.ServerErrorResponse(w, r, err, s.logger)
			}
			return
		}

		r = user.SetUserContext(r, u)

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := user.GetUserContext(r)

		if u.IsAnonymous() {
			jsonutil.AuthenticationRequireResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *Server) requireVerification(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		u := user.GetUserContext(r)
		if !u.Verified {
			jsonutil.VerificationRequireResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	}

	return s.requireAuthentication(fn)
}

func (s *Server) requireRole(role user.Role, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		u := user.GetUserContext(r)

		if u.Role != role {
			jsonutil.InvalidCredentialsResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	}

	return s.requireVerification(fn)
}
