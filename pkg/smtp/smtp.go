package smtp_client

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/gomail.v2"
)

const htmlTemplateDir = "files/html"

type smtpClient struct {
	dialer    *gomail.Dialer
	from      string
	templates *template.Template
	send      func(message *gomail.Message) error
}

func NewSmtpClient(host string, port string, username string, password string, from string) (SmtpClient, error) {
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	templates, err := loadHTMLTemplates(htmlTemplateDir)
	if err != nil {
		return nil, err
	}

	return &smtpClient{
		dialer:    gomail.NewDialer(host, portInt, username, password),
		from:      from,
		templates: templates,
	}, nil
}

var _ SmtpClient = (*smtpClient)(nil)

func (s *smtpClient) SendEmail(email string, subject string, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", s.from)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)
	return s.deliver(message)
}

func (s *smtpClient) SendHTMLTemplate(email string, subject string, templateName string, data any) error {
	body, err := s.renderHTML(templateName, data)
	if err != nil {
		return err
	}
	return s.SendEmail(email, subject, body)
}

func (s *smtpClient) deliver(message *gomail.Message) error {
	if s.send != nil {
		return s.send(message)
	}
	return s.dialer.DialAndSend(message)
}

func loadHTMLTemplates(dir string) (*template.Template, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("smtp: html template path %s is not a directory", dir)
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, nil
	}
	templates, err := template.ParseFiles(matches...)
	if err != nil {
		return nil, fmt.Errorf("smtp: parse html templates in %s: %w", dir, err)
	}
	return templates, nil
}

func (s *smtpClient) renderHTML(templateName string, data any) (string, error) {
	if templateName == "" {
		return "", fmt.Errorf("smtp: html template name is required")
	}
	if s.templates == nil {
		return "", fmt.Errorf("smtp: html templates are not loaded")
	}

	name := templateName
	if filepath.Ext(name) != ".html" {
		name += ".html"
	}

	var body bytes.Buffer
	if err := s.templates.ExecuteTemplate(&body, name, data); err != nil {
		return "", fmt.Errorf("smtp: render html template %s: %w", name, err)
	}
	return body.String(), nil
}
