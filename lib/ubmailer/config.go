package ubmailer

import "fmt"

type MailerType string

const (
	// No mailer is used
	None MailerType = "none"

	// Mailer that does nothing
	Noop MailerType = "Noop"

	// Writes emails to a file - does not send
	File MailerType = "file"

	// Sends emails using an SMTP server
	SMTP MailerType = "smtp"
)

type MailerConfig struct {
	Type MailerType

	// Default from address
	From string

	// File mailer
	OutputDir string

	// SMTP mailer
	Username string
	Password string
	Host     string
}

func MaybeNewMailer(config MailerConfig) Mailer {
	switch config.Type {
	case File:
		if config.From == "" {
			panic("file mailer requires from address")
		}
		if config.OutputDir == "" {
			panic("file mailer requires output directory")
		}
		return NewFileMailer(config.From, config.OutputDir)
	case SMTP:
		if config.Username == "" {
			panic("smtp mailer requires username")
		}
		if config.Password == "" {
			panic("smtp mailer requires password")
		}
		if config.Host == "" {
			panic("smtp mailer requires host")
		}
		if config.From == "" {
			panic("smtp mailer requires from address")
		}
		smtpConfig := SMTPConfig{
			Username: config.Username,
			Password: config.Password,
			Host:     config.Host,
			From:     config.From,
		}

		mailer, err := NewSMTPMailer(smtpConfig)
		if err != nil {
			panic(fmt.Errorf("failed to create smtp mailer: %w", err))
		}
		return mailer
	case None:
		return nil
	case Noop:
		return NewNoopMailer()
	default:
		panic(fmt.Errorf("unknown mailer type: %s", config.Type))
	}
}
