package ubmailer

type NoopMailer struct {
}

func NewNoopMailer() Mailer {
	return &NoopMailer{}
}

func (n *NoopMailer) Send(job EmailJob) error {
	return nil
}
