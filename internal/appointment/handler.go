package appointment

import (
	"appointments/internal/jsonutil"
	"appointments/internal/user"
	"appointments/internal/validator"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	store     *Store
	userStore *user.Store
	logger    *slog.Logger
}

func NewHandler(store *Store, userStore *user.Store, logger *slog.Logger) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
		logger:    logger,
	}
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	apt, ok := h.getAppointment(w, r)
	if !ok {
		return
	}

	err := jsonutil.WriteJSON(w, http.StatusOK, apt, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request, status Status) {
	apt, ok := h.getAppointment(w, r)
	if !ok {
		return
	}
	v := validator.New()

	if ValidateStatus(v, apt.Status, status); !v.Valid() {
		jsonutil.FailedValidationResponse(w, v.Errors)
		return
	}

	apt.Status = status

	h.updateAppointment(w, r, apt)
}

func (h *Handler) updateData(w http.ResponseWriter, r *http.Request) {
	apt, ok := h.getAppointment(w, r)
	if !ok {
		return
	}
	var input struct {
		ProviderID  *int64     `json:"provider_id"`
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		StartsAt    *time.Time `json:"starts_at"`
		EndsAt      *time.Time `json:"ends_at"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		jsonutil.BadRequestResponse(w, err)
		return
	}

	if input.ProviderID != nil {
		apt.ProviderID = *input.ProviderID
	}

	if input.Title != nil {
		apt.Title = *input.Title
	}

	if input.Description != nil {
		apt.Description = *input.Description
	}

	if input.StartsAt != nil {
		apt.StartsAt = *input.StartsAt
	}

	if input.EndsAt != nil {
		apt.EndsAt = *input.EndsAt
	}

	v := validator.New()

	if ValidateAppointment(v, apt); !v.Valid() {
		jsonutil.FailedValidationResponse(w, v.Errors)
		return
	}

	h.updateAppointment(w, r, apt)
}

func (h *Handler) updateAppointment(w http.ResponseWriter, r *http.Request, apt *Appointment) {
	err := h.store.Update(r.Context(), apt)
	if err != nil {
		switch {
		case errors.Is(err, ErrEditConflict):
			jsonutil.EditConflictResponse(w)
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, apt, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}

func (h *Handler) getAppointment(w http.ResponseWriter, r *http.Request) (*Appointment, bool) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		jsonutil.BadRequestResponse(w, err)
		return nil, false
	}

	apt, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrAppointmentNotFound):
			jsonutil.NotFoundResponse(w)
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
		}
		return nil, false
	}

	ok := h.checkAccessByContext(r, apt)
	if !ok {
		jsonutil.NotFoundResponse(w)
		return nil, false
	}

	return apt, true
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
