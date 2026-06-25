package appointment

import "log/slog"

type Handler struct {
	store  *Store
	logger *slog.Logger
}

func NewHandler(store *Store, logger *slog.Logger) *Handler {
	return &Handler{
		store:  store,
		logger: logger,
	}
}
