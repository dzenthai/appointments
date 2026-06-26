package appointment

import (
	"appointments/internal/jsonutil"
	"appointments/internal/user"
	"errors"
	"log/slog"
	"net/http"
)

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

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		jsonutil.BadRequestResponse(w, err)
		return
	}
	apt, err := h.store.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrAppointmentNotFound):
			jsonutil.NotFoundResponse(w)
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}

	ok := h.checkAccessByContext(r, apt)

	if !ok {
		jsonutil.NotFoundResponse(w)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, apt, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}

func (h *Handler) checkAccessByContext(r *http.Request, apt *Appointment) bool {
	u := user.GetUserContext(r)
	switch {
	case u.Role == user.RoleClient && apt.ClientID == u.ID:
		return true
	case u.Role == user.RoleProvider && apt.ProviderID == u.ID:
		return true
	case u.Role == user.RoleAdmin:
		return true
	default:
		h.logger.Warn("unauthorized appointment access attempt",
			"user_id", u.ID,
			"user_role", u.Role,
			"appointment_id", apt.ID,
			"appointment_client", apt.ClientID,
			"appointment_provider", apt.ProviderID,
		)
		return false
	}
}
