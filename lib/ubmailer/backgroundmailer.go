package ubmailer

import (
	"fmt"
	"log/slog"
)

type BackgroundMailer struct {
	mailer Mailer
	queue  chan EmailJob
	done   chan struct{}
}

func NewBackgroundMailer(m Mailer) *BackgroundMailer {
	return &BackgroundMailer{
		mailer: m,
		queue:  make(chan EmailJob, 100),
		done:   make(chan struct{}),
	}
}

func (b *BackgroundMailer) Start() {
	go func() {
		for {
			select {
			case job := <-b.queue:
				slog.Info("Sending email", "to", job.To, "subject", job.Subject)
				err := b.mailer.Send(job)
				if err != nil {
					fmt.Printf("Failed to send email: %v\n", err)
				}
			case <-b.done:
				return
			}
		}
	}()
}

func (b *BackgroundMailer) Stop() {
	close(b.done)
}

func (b *BackgroundMailer) Send(job EmailJob) {
	b.queue <- job
}
