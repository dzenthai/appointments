package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/resend/resend-go/v3"
)

//go:embed "templates"
var tmpl embed.FS

type Mailer struct {
	Client *resend.Client
	Sender string
}

func New(apiKey string, sender string) *Mailer {
	return &Mailer{
		Client: resend.NewClient(apiKey),
		Sender: sender,
	}
}

func renderTemplate(tmplFile string, data any) (string, error) {
	ts, err := template.ParseFS(tmpl, fmt.Sprintf("templates/%s", tmplFile))
	if err != nil {
		return "", err
	}

	var html bytes.Buffer

	if err = ts.ExecuteTemplate(&html, "html", data); err != nil {
		return "", err
	}

	return html.String(), nil
}

func (m *Mailer) send(subject, html, recipient string, logger *slog.Logger) error {

	client := m.Client

	params := &resend.SendEmailRequest{
		From:    m.Sender,
		To:      []string{recipient},
		Subject: subject,
		Html:    html,
	}

	response, err := client.Emails.Send(params)
	if err != nil {
		logger.Error("failed to send email", "err", err)
		return err
	}
	logger.Info("email sent", "id", response.Id)
	return nil
}
