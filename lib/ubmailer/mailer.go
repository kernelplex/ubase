package ubmailer

import (
	"fmt"
	"github.com/wneessen/go-mail"
)

type SMTPConfig struct {
	Username string
	Password string
	Host     string
	From     string
}

type MailerImpl struct {
	client *mail.Client
	from   string
}

type EmailJob struct {
	To       string
	Subject  string
	TextBody string
	HtmlBody string
}

type Mailer interface {
	Send(job EmailJob) error
}

func NewSMTPMailer(cfg SMTPConfig) (Mailer, error) {
	client, err := mail.NewClient(
		cfg.Host,
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	return &MailerImpl{
		client: client,
		from:   cfg.From,
	}, nil
}

func (m *MailerImpl) Send(job EmailJob) error {
	message := mail.NewMsg()

	if err := message.From(m.from); err != nil {
		return fmt.Errorf("failed to set from: %w", err)
	}

	if err := message.To(job.To); err != nil {
		return fmt.Errorf("failed to set to: %w", err)
	}

	message.Subject(job.Subject)
	if len(job.HtmlBody) > 0 {
		message.SetBodyString(mail.TypeTextHTML, job.HtmlBody)
		message.AddAlternativeString(mail.TypeTextPlain, job.TextBody)
	} else {
		message.SetBodyString(mail.TypeTextPlain, job.TextBody)
	}

	if err := m.client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
