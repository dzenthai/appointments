package mailer

import (
	"log/slog"
)

type CodeEmailData struct {
	Subject string
	Message string
	Hint    string
	Code    string
}

func (m *Mailer) SendVerification(email, code string, logger *slog.Logger) error {
	data := CodeEmailData{
		Subject: "Email Verification",
		Message: "To complete your registration, please use the confirmation code:",
		Hint:    "If you did not request registration, ignore this email.",
		Code:    code,
	}

	html, err := renderTemplate("email_verification.tmpl", data)
	if err != nil {
		return err
	}

	return m.send(data.Subject, html, email, logger)
}
