package mailer

import (
	"net/http"

	"google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/mail"
)

// Mailer is an empty struct.
type Mailer struct{}

// NewMailer returns a new instance of the Mailer struct.
func NewMailer() *Mailer {
	return &Mailer{}
}

// SendEmail sends an email using the App Engine mail service.
func (m *Mailer) SendEmail(r *http.Request, sender string, recipient []string, subject string, body string) error {
	ctx := appengine.NewContext(r)
	msg := &mail.Message{
		Sender:  sender,
		To:      recipient,
		Subject: subject,
		Body:    body,
	}
	if err := mail.Send(ctx, msg); err != nil {
		return err
	}
	return nil
}
