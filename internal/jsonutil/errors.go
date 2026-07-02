package jsonutil

import (
	"log/slog"
	"net/http"
)

func errorResponse(w http.ResponseWriter, status int, message any) {
	ed := map[string]any{
		"error": message,
	}

	_ = WriteJSON(w, status, ed, nil)
}

func ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	logger.Error("error occurs", "err", err, "method", r.Method, "uri", r.RequestURI)
	message := "the server encountered a problem and could not process your request"
	errorResponse(w, http.StatusInternalServerError, message)
}

func InvalidVerificationTokenResponse(w http.ResponseWriter) {
	message := "invalid verification token"
	errorResponse(w, http.StatusNotFound, message)
}

func BadRequestResponse(w http.ResponseWriter, err error) {
	errorResponse(w, http.StatusBadRequest, err.Error())
}

func EditConflictResponse(w http.ResponseWriter) {
	message := "unable to update the record due to an edit conflict, please try again"
	errorResponse(w, http.StatusConflict, message)
}

func FailedValidationResponse(w http.ResponseWriter, errors map[string]string) {
	errorResponse(w, http.StatusBadRequest, errors)
}

func InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	errorResponse(w, http.StatusUnauthorized, message)
}

func InvalidCredentialsResponse(w http.ResponseWriter) {
	message := "invalid credentials"
	errorResponse(w, http.StatusUnauthorized, message)
}

func AuthenticationRequireResponse(w http.ResponseWriter) {
	message := "authentication required"
	errorResponse(w, http.StatusUnauthorized, message)
}

func VerificationRequireResponse(w http.ResponseWriter) {
	message := "verification required"
	errorResponse(w, http.StatusForbidden, message)
}

func NotFoundResponse(w http.ResponseWriter) {
	message := "not found"
	errorResponse(w, http.StatusNotFound, message)
}

func LimitExceededResponse(w http.ResponseWriter) {
	message := "too many requests, limit exceeded"
	errorResponse(w, http.StatusTooManyRequests, message)
}
