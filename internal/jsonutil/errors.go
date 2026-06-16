package jsonutil

import (
	"log/slog"
	"net/http"
)

func errorResponse(w http.ResponseWriter, r *http.Request, status int, message any, logger *slog.Logger) {
	ed := map[string]any{
		"error": message,
	}

	err := WriteJSON(w, status, ed, nil)
	if err != nil {
		logError(r, err, logger)
		return
	}
}

func logError(r *http.Request, err error, logger *slog.Logger) {
	logger.Error("error occurs", "err", err, "method", r.Method, "uri", r.RequestURI)
}

func serverErrorResponse(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	logError(r, err, logger)
	message := "the server encountered a problem and could not process your request"
	errorResponse(w, r, http.StatusInternalServerError, message, logger)
}

func notFoundResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "the requested resource could not be found"
	errorResponse(w, r, http.StatusNotFound, message, logger)
}

func badRequestResponse(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	errorResponse(w, r, http.StatusBadRequest, err.Error(), logger)
}
