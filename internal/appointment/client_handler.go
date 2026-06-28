package appointment

import (
	"appointments/internal/jsonutil"
	"appointments/internal/user"
	"appointments/internal/validator"
	"errors"
	"net/http"
	"time"
)

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

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	h.updateData(w, r)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, StatusCancelled)
}
