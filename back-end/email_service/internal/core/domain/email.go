package domain

type EmailContent struct {
	Recipient string `json:"recipient"`
	Subject string `json:"subject"`
	Body string `json:"body"`
}
