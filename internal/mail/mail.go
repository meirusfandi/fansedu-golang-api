package mail

import "fmt"

// Mailer mengirim email. Implementasi noop hanya log; production bisa pakai SMTP/SendGrid.
type Mailer interface {
	Send(to, subject, bodyPlain string) error
}

// LogMailer hanya menulis ke log (dev); tidak mengirim email sungguhan.
type LogMailer struct{}

func (m *LogMailer) Send(to, subject, bodyPlain string) error {
	fmt.Printf("[mail] to=%q subject=%q body_len=%d\n%s\n", to, subject, len(bodyPlain), bodyPlain)
	return nil
}

// NewLogMailer returns a mailer that logs instead of sending (for development).
func NewLogMailer() *LogMailer { return &LogMailer{} }
