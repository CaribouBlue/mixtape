package mail

import (
	"github.com/wneessen/go-mail"
)

type GmailMailer struct {
	*mail.Client
	fromAddress string
}

func NewGmailMailer(username, password string) (*GmailMailer, error) {
	client, err := mail.NewClient("smtp.gmail.com", mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(username), mail.WithPassword(password))
	if err != nil {
		return nil, err
	}

	return &GmailMailer{
		client,
		username,
	}, nil
}

func (m *GmailMailer) SendMail(to, subject, body string) error {
	message := mail.NewMsg()

	err := message.From(m.fromAddress)
	if err != nil {
		return err
	}

	err = message.To(to)
	if err != nil {
		return ErrInvalidEmail
	}

	message.Subject(subject)
	message.SetBodyString(mail.TypeTextPlain, body)

	err = m.DialAndSend(message)
	if err != nil {
		return err
	}

	return nil
}
