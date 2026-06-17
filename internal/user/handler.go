package user

import (
	"appointments/internal/jsonutil"
	"appointments/internal/validator"
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

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName         string `json:"first_name"`
		SecondName        string `json:"second_name"`
		Email             string `json:"email"`
		PlaintextPassword string `json:"password"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		jsonutil.BadRequestResponse(w, err)
		return
	}

	user := User{
		FirstName:  input.FirstName,
		SecondName: input.SecondName,
		Email:      input.Email,
	}

	err = user.Password.Set(input.PlaintextPassword)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	v := validator.New()

	if ValidateUser(v, user); !v.Valid() {
		jsonutil.FailedValidationResponse(w, v.Errors)
		return
	}

	err = h.store.Insert(&user)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			jsonutil.FailedValidationResponse(w, v.Errors)
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusAccepted, user, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}

}
