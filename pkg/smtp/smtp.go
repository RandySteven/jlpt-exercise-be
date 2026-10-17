package smtp_client

import "gopkg.in/gomail.v2"

type smtpClient struct {
	dialer   *gomail.Dialer
	message  *gomail.Message
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSmtpClient(host string, port string, username string, password string, from string) SmtpClient {
	dialer := gomail.NewDialer(host, port, username, password)
	message := gomail.NewMessage()

	return &smtpClient{
		dialer:   dialer,
		message:  message,
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *smtpClient) SendEmail(email string, subject string, body string) error {
	s.message.SetHeader("From", s.from)
	s.message.SetHeader("To", email)
	s.message.SetHeader("Subject", subject)
	s.message.SetBody("text/html", body)

	return s.dialer.DialAndSend(s.message)
}
