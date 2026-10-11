package service

import (
	"encoding/json"
	"fmt"
	"hoshiBmaTchi/back-end/email_service/internal/core/domain"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	dialer     *gomail.Dialer
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewEmailService() *EmailService {
	dialer := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		587,
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
	)

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://rabbitmq:rabbitmq@rabbitmq:5672/"
	}

	var conn *amqp.Connection
	var err error
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ")
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ after %d attempts: %v. Email service will not queue messages.", maxRetries, err)
		return &EmailService{dialer: dialer, connection: nil, channel: nil}
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open RabbitMQ channel: %v", err)
		conn.Close()
		return &EmailService{dialer: dialer, connection: nil, channel: nil}
	}

	_, err = ch.QueueDeclare("queue", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		ch.Close()
		conn.Close()
		return &EmailService{dialer: dialer, connection: nil, channel: nil}
	}

	return &EmailService{
		dialer:     dialer,
		connection: conn,
		channel:    ch,
	}
}

func (es *EmailService) QueueOTPEmail(recipient string, code string) error {
	body := fmt.Sprintf(`
    <h1>OTP Code</h1>
    <p>Your OTP code is:</p>
    <h2><b>%s</b></h2>
    <p>This code will expire in 5 minutes.</p>
    `, code)

	msg := domain.EmailContent{
		Recipient: recipient,
		Subject:   "OTP Verification Code",
		Body:      body,
	}

	if es.channel == nil {
		log.Printf("Skipping OTP email queue - RabbitMQ not available")
		return fmt.Errorf("rabbitmq channel is not available")
	}

	data, _ := json.Marshal(msg)
	return es.channel.Publish(
		"", "queue", false, false, amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         data,
		},
	)
}

func (es *EmailService) QueueWelcomeEmail(recipient string, name string) error {
	body := fmt.Sprintf(`
        <h1>Welcome to Hoshibmatchi, %s!</h1>
        <p>Your account has been successfully verified.</p>
    `, name)

	msg := domain.EmailContent{
		Recipient: recipient,
		Subject:   "Hoshibmatchi Registration Succeeded",
		Body:      body,
	}

	if es.channel == nil {
		log.Printf("Skipping welcome email queue - RabbitMQ not available")
		return fmt.Errorf("rabbitmq channel is not available")
	}

	data, _ := json.Marshal(msg)
	return es.channel.Publish(
		"", "queue", false, false, amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         data,
		},
	)
}

func (es *EmailService) QueueLoginEmail(recipient string, name string) error {
	body := fmt.Sprintf(`
        <h3>You have successfully logged in</h3>
        <h1>Welcome, %s!</h1>
    `, name)

	msg := domain.EmailContent{
		Recipient: recipient,
		Subject:   "Hoshibmatchi Login Succeeded",
		Body:      body,
	}

	if es.channel == nil {
		log.Printf("Skipping login email queue - RabbitMQ not available")
		return fmt.Errorf("rabbitmq channel is not available")
	}

	data, _ := json.Marshal(msg)
	return es.channel.Publish(
		"", "queue", false, false, amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         data,
		},
	)
}

func (es *EmailService) EmailWorker() {
	if es.channel == nil {
		log.Println("Cannot start EmailWorker: RabbitMQ channel is nil")
		return
	}

	msgs, err := es.channel.Consume("queue", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	smtpUser := os.Getenv("SMTP_USER")

	go func() {
		for m := range msgs {
			var email domain.EmailContent
			err := json.Unmarshal(m.Body, &email)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			mail := gomail.NewMessage()
			mail.SetHeader("From", smtpUser)
			mail.SetHeader("To", email.Recipient)
			mail.SetHeader("Subject", email.Subject)
			mail.SetBody("text/html", email.Body)

			if err := es.dialer.DialAndSend(mail); err != nil {
				log.Printf("Failed to send email to %s: %v", email.Recipient, err)
			} else {
				log.Printf("Email sent successfully to %s", email.Recipient)
			}
		}
	}()
	log.Println("EmailWorker is running and waiting for messages...")
}