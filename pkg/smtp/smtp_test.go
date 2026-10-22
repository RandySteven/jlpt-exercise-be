package smtp_client

import (
	"bytes"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/gomail.v2"
)

func TestLoadHTMLTemplatesMissingDir(t *testing.T) {
	templates, err := loadHTMLTemplates(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatalf("load missing dir: %v", err)
	}
	if templates != nil {
		t.Fatal("missing html folder should leave templates unloaded")
	}
}

func TestLoadHTMLTemplatesParsesHTMLFiles(t *testing.T) {
	dir := t.TempDir()
	writeHTML(t, dir, "welcome.html", `<p>Hello {{.Name}}</p>`)
	writeHTML(t, dir, "note.txt", `not a template`)

	templates, err := loadHTMLTemplates(dir)
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}
	if templates == nil || templates.Lookup("welcome.html") == nil {
		t.Fatal("welcome.html was not loaded")
	}
	if templates.Lookup("note.txt") != nil {
		t.Fatal("non-html files should be ignored")
	}
}

func TestLoadHTMLTemplatesRejectsInvalidTemplate(t *testing.T) {
	dir := t.TempDir()
	writeHTML(t, dir, "broken.html", `<p>{{.Name</p>`)

	if _, err := loadHTMLTemplates(dir); err == nil {
		t.Fatal("invalid html template should fail to parse")
	}
}

func TestSendHTMLTemplateRendersFolderTemplate(t *testing.T) {
	dir := t.TempDir()
	writeHTML(t, dir, "welcome.html", `<p>Hello {{.Name}}</p>`)

	templates, err := loadHTMLTemplates(dir)
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	var sent *gomail.Message
	client := &smtpClient{
		from:      "jlpt@example.com",
		templates: templates,
		send: func(message *gomail.Message) error {
			sent = message
			return nil
		},
	}

	err = client.SendHTMLTemplate("ada@example.com", "Welcome", "welcome", map[string]string{"Name": "Ada"})
	if err != nil {
		t.Fatalf("send template: %v", err)
	}
	if sent == nil {
		t.Fatal("template email was not sent")
	}
	if got := sent.GetHeader("From"); len(got) != 1 || got[0] != "jlpt@example.com" {
		t.Fatalf("from = %#v", got)
	}
	if got := sent.GetHeader("To"); len(got) != 1 || got[0] != "ada@example.com" {
		t.Fatalf("to = %#v", got)
	}
	if got := sent.GetHeader("Subject"); len(got) != 1 || got[0] != "Welcome" {
		t.Fatalf("subject = %#v", got)
	}
	if body := messageHTML(t, sent); body != "<p>Hello Ada</p>" {
		t.Fatalf("body = %q", body)
	}
}

func TestSendHTMLTemplateWithoutFolder(t *testing.T) {
	client := &smtpClient{}
	err := client.SendHTMLTemplate("ada@example.com", "Welcome", "welcome", nil)
	if err == nil || !strings.Contains(err.Error(), "html templates are not loaded") {
		t.Fatalf("err = %v", err)
	}
}

func TestNewSmtpClientRejectsInvalidPort(t *testing.T) {
	client, err := NewSmtpClient("localhost", "mail", "user", "pass", "jlpt@example.com")
	if err == nil || client != nil {
		t.Fatalf("client = %#v err = %v", client, err)
	}
}

func writeHTML(t *testing.T, dir string, name string, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func messageHTML(t *testing.T, message *gomail.Message) string {
	t.Helper()
	var raw bytes.Buffer
	if _, err := message.WriteTo(&raw); err != nil {
		t.Fatalf("write message: %v", err)
	}
	parsed, err := mail.ReadMessage(&raw)
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	body, err := io.ReadAll(quotedprintable.NewReader(parsed.Body))
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return strings.TrimSpace(string(body))
}
