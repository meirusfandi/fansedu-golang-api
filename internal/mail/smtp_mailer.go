package mail

import (
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPConfig untuk relay SMTP (mis. Brevo: smtp-relay.brevo.com:587).
// User: biasanya email akun Brevo yang terverifikasi; kosong = pakai From.
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// SMTPMailer mengirim email teks biasa lewat SMTP + STARTTLS (port 587).
type SMTPMailer struct {
	host, user, password, from string
	port                       int
}

// NewSMTPMailer memvalidasi konfigurasi minimal.
func NewSMTPMailer(c SMTPConfig) (*SMTPMailer, error) {
	if strings.TrimSpace(c.Host) == "" || strings.TrimSpace(c.Password) == "" || strings.TrimSpace(c.From) == "" {
		return nil, fmt.Errorf("mail: smtp host, password, and from are required")
	}
	port := c.Port
	if port <= 0 {
		port = 587
	}
	user := strings.TrimSpace(c.User)
	if user == "" {
		user = strings.TrimSpace(c.From)
	}
	return &SMTPMailer{
		host:     strings.TrimSpace(c.Host),
		port:     port,
		user:     user,
		password: c.Password,
		from:     strings.TrimSpace(c.From),
	}, nil
}

// Send mengirim satu pesan text/plain UTF-8.
func (m *SMTPMailer) Send(to, subject, bodyPlain string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("mail: empty recipient")
	}
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.password, m.host)

	var b strings.Builder
	b.WriteString("From: ")
	b.WriteString(m.from)
	b.WriteString("\r\nTo: ")
	b.WriteString(to)
	b.WriteString("\r\nSubject: ")
	b.WriteString(subject)
	b.WriteString("\r\nMIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(bodyPlain)

	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(b.String()))
}
