package mailer

import (
	"log/slog"
)

type VerificationEmailData struct {
	Subject    string
	FirstName  string
	SecondName string
	Message    string
	Hint       string
	Code       string
}

func (m *Mailer) SendVerification(firstName, secondName, email, code string, logger *slog.Logger) error {
	data := VerificationEmailData{
		Subject:    "Email Confirmation",
		FirstName:  firstName,
		SecondName: secondName,
		Message:    "To complete your registration, please use the confirmation code:",
		Hint:       "If you did not request registration, ignore this email.",
		Code:       code,
	}

	html, err := renderTemplate("email_confirmation.tmpl", data)
	if err != nil {
		return err
	}

	return m.send("Email confirmation", html, email, logger)
}
