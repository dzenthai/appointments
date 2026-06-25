package user

import (
	"appointments/internal/background"
	"appointments/internal/jsonutil"
	"appointments/internal/mailer"
	"appointments/internal/token"
	"appointments/internal/validator"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Handler struct {
	store        *Store
	token        *token.Store
	logger       *slog.Logger
	wg           *sync.WaitGroup
	mailer       *mailer.Mailer
	vryTokenTTL  time.Duration
	authTokenTTL time.Duration
}

func NewHandler(
	store *Store,
	token *token.Store,
	logger *slog.Logger,
	wg *sync.WaitGroup,
	mailer *mailer.Mailer,
	vryTokenTTL time.Duration,
	authTokenTTL time.Duration,
) *Handler {
	return &Handler{
		store:        store,
		token:        token,
		logger:       logger,
		wg:           wg,
		mailer:       mailer,
		vryTokenTTL:  vryTokenTTL,
		authTokenTTL: authTokenTTL,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName         string `json:"first_name"`
		SecondName        string `json:"second_name"`
		Email             string `json:"email"`
		PlaintextPassword string `json:"password"`
		Role              string `json:"role"`
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
		Role:       Role(input.Role),
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
				h.sendExistingAccount(*existing)
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

	h.sendVerificationCode(user)

	err = jsonutil.WriteJSON(w, http.StatusAccepted, jsonutil.Envelope{"message": "check your email to complete registration"}, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}

func (h *Handler) sendVerificationCode(user User) {
	background.Run(h.wg, h.logger, func() {
		vry, err := token.NewVerification(user.ID, h.vryTokenTTL)
		if err != nil {
			h.logger.Error("failed to create verification token", "err", err)
			return
		}

		err = h.token.CreateVerification(vry)
		if err != nil {
			h.logger.Error("failed to save verification token", "err", err)
			return
		}

		err = h.mailer.SendVerification(user.Email, vry.Plaintext)
		if err != nil {
			h.logger.Error("failed to send verification email", "err", err)
			return
		}
	})
}

func (h *Handler) sendExistingAccount(user User) {
	background.Run(h.wg, h.logger, func() {
		err := h.mailer.SendExistingAccount(user.Email)
		if err != nil {
			h.logger.Error("failed to send existing account email", "err", err)
		}
	})
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
			jsonutil.InvalidVerificationTokenResponse(w)
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

	err = h.token.DeleteVerificationsByUserID(user.ID)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, user, nil)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email             string `json:"email"`
		PlaintextPassword string `json:"password"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		jsonutil.BadRequestResponse(w, err)
		return
	}

	user, err := h.store.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			jsonutil.InvalidCredentialsResponse(w)
		default:
			jsonutil.ServerErrorResponse(w, r, err, h.logger)
		}
		return
	}

	match, err := user.Password.Matches(input.PlaintextPassword)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	if !match {
		jsonutil.InvalidCredentialsResponse(w)
		return
	}

	authToken, err := token.NewAuthentication(user.ID, h.authTokenTTL)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = h.token.CreateAuthentication(authToken)
	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
		return
	}

	err = jsonutil.WriteJSON(
		w,
		http.StatusCreated,
		jsonutil.Envelope{
			"token":      authToken.Plaintext,
			"expires_at": authToken.ExpiresAt,
		},
		nil,
	)

	if err != nil {
		jsonutil.ServerErrorResponse(w, r, err, h.logger)
	}
}
