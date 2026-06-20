package user

import (
	"appointments/internal/jsonutil"
	"appointments/internal/mailer"
	"appointments/internal/token"
	"appointments/internal/validator"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	store       *Store
	token       *token.Store
	logger      *slog.Logger
	mailer      *mailer.Mailer
	vryTokenTTL time.Duration
}

func NewHandler(store *Store, token *token.Store, logger *slog.Logger, mailer *mailer.Mailer, vryTokenTTL time.Duration) *Handler {
	return &Handler{
		store:       store,
		token:       token,
		logger:      logger,
		mailer:      mailer,
		vryTokenTTL: vryTokenTTL,
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
			if existing.Verified {
				err = jsonutil.WriteJSON(w, http.StatusAccepted, jsonutil.Envelope{"message": "check your email to complete registration"}, nil)
				if err != nil {
					jsonutil.ServerErrorResponse(w, r, err, h.logger)
				}
				return
			}
			user = *existing
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
			return
		}
	}

	if err = h.sendVerificationCode(user); err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusAccepted, jsonutil.Envelope{"message": "check your email to complete registration"}, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}

func (h *Handler) sendVerificationCode(user User) error {
	vry, err := token.NewVerification(user.ID, h.vryTokenTTL)
	if err != nil {
		return err
	}

	err = h.token.Create(vry)
	if err != nil {
		return err
	}

	return h.mailer.SendVerification(user.Email, vry.Plaintext, h.logger)
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

	user, err := h.store.GetByToken(input.Plaintext, token.ScopeVerification)
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

	err = h.token.DeleteAllByUserID(user.ID)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, user, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}
