package ubmailer

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/wneessen/go-mail"
)

type FileMailer struct {
	from      string
	outputDir string
}

func NewFileMailer(from string, outputDir string) *FileMailer {
	return &FileMailer{
		from:      from,
		outputDir: outputDir,
	}
}

func (f *FileMailer) Send(job EmailJob) error {
	slog.Info("Sending email", "to", job.To, "subject", job.Subject)
	m := mail.NewMsg()
	if err := m.From(f.from); err != nil {
		return fmt.Errorf("failed to set from: %w", err)
	}
	if err := m.To(job.To); err != nil {
		return fmt.Errorf("failed to set to: %w", err)
	}

	m.Subject(job.Subject)
	if len(job.HtmlBody) > 0 {
		m.SetBodyString(mail.TypeTextHTML, job.HtmlBody)
		m.AddAlternativeString(mail.TypeTextPlain, job.TextBody)
	} else {
		m.SetBodyString(mail.TypeTextPlain, job.TextBody)
	}

	// Ensure output directory exists
	if f.outputDir != "" {
		if err := os.MkdirAll(f.outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	filename := fmt.Sprintf("email_%s_%s.eml", time.Now().Format("20060102_150405"), randomSlug())
	filePath := filename
	if f.outputDir != "" {
		filePath = fmt.Sprintf("%s/%s", f.outputDir, filename)
	}
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := m.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write email to file: %w", err)
	}

	return nil
}

func randomSlug() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
