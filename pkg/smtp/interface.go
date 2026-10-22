package smtp_client

type (
	SmtpClient interface {
		SendEmail(email string, subject string, body string) error
		SendHTMLTemplate(email string, subject string, templateName string, data any) error
	}
)
