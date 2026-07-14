package appointment

import (
	"appointments/internal/filters"
	"appointments/internal/httputil"
	"appointments/internal/jsonutil"
	"appointments/internal/user"
	"appointments/internal/validator"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	store     appointmentStore
	userStore *user.Store
	logger    *slog.Logger
}

type appointmentStore interface {
	GetAllByAdmin(ctx context.Context, f filters.Filters) ([]Appointment, error)
	GetAllByClient(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error)
	GetAllByProvider(ctx context.Context, userID int64, f filters.Filters) ([]Appointment, error)
	GetByID(ctx context.Context, id int64) (*Appointment, error)
	Insert(ctx context.Context, apt *Appointment) error
	Update(ctx context.Context, apt *Appointment) error
}

func NewHandler(store appointmentStore, userStore *user.Store, logger *slog.Logger) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
		logger:    logger,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {

	qs := r.URL.Query()

	v := validator.New()

	fs := filters.Filters{
		Page:         httputil.ReadInt(qs, "page", 1, v),
		PageSize:     httputil.ReadInt(qs, "page_size", 20, v),
		Sort:         httputil.ReadString(qs, "sort", "title"),
		SortSafeList: []string{"title", "-title", "starts_at", "-starts_at", "ends_at", "-ends_at", "status", "-status"},
	}

	if filters.ValidateFilters(v, fs); !v.Valid() {
		jsonutil.FailedValidationResponse(w, v.Errors)
		return
	}

	u := user.GetUserContext(r)

	var apts []Appointment
	var err error

	switch u.Role {
	case user.RoleClient:
		apts, err = h.store.GetAllByClient(r.Context(), u.ID, fs)
	case user.RoleProvider:
		apts, err = h.store.GetAllByProvider(r.Context(), u.ID, fs)
	case user.RoleAdmin:
		apts, err = h.store.GetAllByAdmin(r.Context(), fs)
	default:
		jsonutil.NotFoundResponse(w)
		return
	}

	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, apts, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
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

	status := StatusScheduled

	v := validator.New()

	ValidateStatus(v, apt.Status, status)

	ValidateAppointment(v, apt)

	if !v.Valid() {
		jsonutil.FailedValidationResponse(w, v.Errors)
		return
	}

	apt.Status = status

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
