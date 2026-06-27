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
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		jsonutil.BadRequestResponse(w, err)
		return
	}
	apt, err := h.store.GetByID(r.Context(), id)
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProviderID  int64     `json:"provider_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		StartsAt    time.Time `json:"starts_at"`
		EndsAt      time.Time `json:"ends_at"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		jsonutil.BadRequestResponse(w, err)
		return
	}

	v := validator.New()

	u := user.GetUserContext(r)
	if u.Role != user.RoleClient {
		v.AddError("client_id", "invalid client role")
		jsonutil.FailedValidationResponse(w, v.Errors)
		return
	}

	provider, err := h.userStore.GetByID(r.Context(), input.ProviderID)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserNotFound):
			jsonutil.BadRequestResponse(w, err)
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}
	if provider.Role != user.RoleProvider {
		v.AddError("provider_id", "invalid provider role")
		jsonutil.FailedValidationResponse(w, v.Errors)
		return
	}

	apt := &Appointment{
		ClientID:    u.ID,
		ProviderID:  provider.ID,
		Title:       input.Title,
		Description: input.Description,
		StartsAt:    input.StartsAt,
		EndsAt:      input.EndsAt,
		Status:      StatusScheduled,
	}

	if ValidateAppointment(v, apt); !v.Valid() {
		jsonutil.FailedValidationResponse(w, v.Errors)
		return
	}

	err = h.store.Insert(r.Context(), apt)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusCreated, apt, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}
}
