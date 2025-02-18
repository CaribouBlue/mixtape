package mail

import "errors"

var (
	ErrInvalidEmail = errors.New("invalid email address")
)

type Mailer interface {
	SendMail(to, subject, body string) error
}

type MailService struct {
	Mailer Mailer
}

func NewMailService(mailer Mailer) *MailService {
	return &MailService{
		Mailer: mailer,
	}
}

func (m *MailService) SendMail(to, subject, body string) error {
	err := m.Mailer.SendMail(to, subject, body)
	if err != nil {
		return err
	}
	return nil
}
