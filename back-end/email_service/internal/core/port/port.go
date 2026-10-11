package port

type EmailServicePort interface {
	QueueOTPEmail(recipient string, code string) error
	QueueWelcomeEmail(recipient, name string) error
	QueueLoginEmail(recipient, name string) error
	EmailWorker()
}