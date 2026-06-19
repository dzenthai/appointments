package user

import (
	"appointments/internal/jsonutil"
	"appointments/internal/mailer"
	"appointments/internal/validator"
	"appointments/internal/verification"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	store         *Store
	verifications *verification.Store
	logger        *slog.Logger
	mailer        *mailer.Mailer
	codeTTL       time.Duration
}

func NewHandler(store *Store, verifications *verification.Store, logger *slog.Logger, mailer *mailer.Mailer, codeTTL time.Duration) *Handler {
	return &Handler{
		store:         store,
		verifications: verifications,
		logger:        logger,
		mailer:        mailer,
		codeTTL:       codeTTL,
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
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateEmail):
			var existing *User
			existing, err = h.store.GetByEmail(input.Email)
			if err != nil {
				jsonutil.ServerErrorResponse(w, r, err, h.logger)
				return
			}
			user = *existing
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
			return
		}
	}

	vry, err := verification.NewCode(user.ID, h.codeTTL)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = h.verifications.Create(vry)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = h.mailer.SendVerification(user.Email, vry.Code.Plaintext(), h.logger)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusAccepted, jsonutil.Envelope{"message": "check your email to complete registration"}, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Plaintext string `json:"code"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		jsonutil.BadRequestResponse(w, err)
		return
	}

	user, err := h.store.GetForCodes(input.Plaintext)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			jsonutil.BadRequestResponse(w, err)
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}

	user.Verified = true

	err = h.store.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateEmail):
			jsonutil.BadRequestResponse(w, err)
		case errors.Is(err, ErrEditConflict):
			jsonutil.EditConflictResponse(w)
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}

	err = h.verifications.DeleteAllByUserID(user.ID)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, user, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}
