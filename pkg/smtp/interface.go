package smtp_client

type (
	SmtpClient interface {
		SendEmail(email string, subject string, body string) error
	}
)
