package mailSender

import "gopkg.in/gomail.v2"

type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
}

func (m *Mailer) Send(to, username, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("To", to)
	msg.SetHeader("From", m.Username)
	msg.SetHeader("Subject", "Подтверждение почты")

	msg.SetBody("text/plain", body)

	dialer := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	return dialer.DialAndSend(msg)
}
