package service

import (
	"fmt"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/config"
	"gopkg.in/gomail.v2"
)

type MailService struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailService() (*MailService, error) {
	dialer, err := config.SMTPConfig()
	if err != nil {
		return nil, err
	}

	return &MailService{
		dialer: dialer,
		from:   "noreply@rivandoalrasyid.it.com",
	}, nil
}

func (m *MailService) SendResetPassword(email, token string) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", m.from)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Reset Your VanWallet Password")

	resetURL := fmt.Sprintf(
		"https://vanwallet.rivandoalrasyid.it.com/reset-password?token=%s",
		token,
	)

	body := fmt.Sprintf(`Hello,

We received a request to reset your VanWallet password.

Click the link below to reset your password:

%s

This link expires in 15 minutes.

If you didn't request this request, you can safely ignore this email.

Regards,
VanWallet Team
`, resetURL)

	msg.SetBody("text/plain", body)

	return m.dialer.DialAndSend(msg)
}

func (m *MailService) SendHTML(to, subject, html string) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", html)

	return m.dialer.DialAndSend(msg)
}

func (m *MailService) SendText(to, subject, body string) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	return m.dialer.DialAndSend(msg)
}
