package user

import (
	"appointments/internal/jsonutil"
	"appointments/internal/mailer"
	"appointments/internal/validator"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	store  *Store
	logger *slog.Logger
	mailer *mailer.Mailer
}

func NewHandler(store *Store, logger *slog.Logger, mailer *mailer.Mailer) *Handler {
	return &Handler{
		store:  store,
		logger: logger,
		mailer: mailer,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
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
	if err != nil && !errors.Is(err, ErrDuplicateEmail) {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = h.mailer.SendVerification(user.FirstName, user.SecondName, user.Email, "blank", h.logger)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusAccepted, jsonutil.Envelope{"message": "check your email to complete registration"}, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}

}
