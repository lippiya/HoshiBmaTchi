package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/grpcclient"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/redis"
	"log"
	"math/big"

	"github.com/streadway/amqp"
)

type OTPService struct {
	redisClient     *redis.RedisClient
	emailGRPCClient *grpcclient.Client
	rabbitChannel   *amqp.Channel
}

func NewOTPService(r *redis.RedisClient, emailClient *grpcclient.Client, rabbitCh *amqp.Channel) *OTPService {
	return &OTPService{
		redisClient:     r,
		emailGRPCClient: emailClient,
		rabbitChannel:   rabbitCh,
	}
}

func GenerateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(999999))
	otp := fmt.Sprintf("%06d", n.Int64())
	return otp
}

func (os *OTPService) SendOTP(c context.Context, email string) (string, error) {
	allowed, ttl, err := os.redisClient.RateLimitCheck(email)
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", fmt.Errorf("please wait %d seconds before requesting a new OTP", ttl)
	}

	code := GenerateOTP()

	err2 := os.redisClient.SetOTP(email, code)
	if err2 != nil {
		return "", err2
	}

	if os.emailGRPCClient != nil {
		if err := os.emailGRPCClient.QueueOTP(c, email, code); err == nil {
			return "OTP sent successfully", nil
		} else {
			log.Printf("gRPC call to QueueOTP failed: %v", err)
		}
	} else {
		log.Printf("Email gRPC client is nil, will try RabbitMQ fallback")
	}

	if os.rabbitChannel != nil {
		body := fmt.Sprintf("<h1>OTP Code</h1><p>Your OTP code is: <b>%s</b></p><p>This code will expire in 5 minutes.</p>", code)
		msg := map[string]string{
			"recipient": email,
			"subject":   "OTP Verification Code",
			"body":      body,
		}
		data, _ := json.Marshal(msg)
		err := os.rabbitChannel.Publish("", "queue", false, false, amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         data,
		})
		if err != nil {
			log.Printf("Failed to publish OTP to RabbitMQ: %v", err)
			return "", fmt.Errorf("failed to send OTP email")
		}
		return "OTP sent successfully (via queue)", nil
	}

	return "", fmt.Errorf("email service is unavailable")
}

func (s *OTPService) VerifyOTP(email string, code string) (bool, error) {
	getCode, err := s.redisClient.GetOTP(email)
	if err != nil {
		return false, fmt.Errorf("OTP expired or not found")
	}

	if getCode != code {
		return false, fmt.Errorf("invalid OTP code")
	}

	s.redisClient.DeleteOTP(email)

	return true, nil
}

func (s *OTPService) SendWelcomeEmail(ctx context.Context, email string, name string) error {

	if s.emailGRPCClient != nil {
		if err := s.emailGRPCClient.QueueWelcome(ctx, email, name); err == nil {
			return nil
		} else {
			log.Printf("gRPC call to QueueWelcomeEmail failed: %v", err)
		}
	}

	if s.rabbitChannel != nil {
		body := fmt.Sprintf("<h1>Welcome to HoshiBmaTchi!</h1><p>Hi %s,</p>", name)
		msg := map[string]string{
			"recipient": email,
			"subject":   "Welcome to HoshiBmaTchi!",
			"body":      body,
		}
		data, _ := json.Marshal(msg)
		err := s.rabbitChannel.Publish("", "queue", false, false, amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body: data,
		})
		if err != nil {
			log.Printf("Failed to publish welcome email to RabbitMQ: %v", err)
			return fmt.Errorf("failed to send welcome email")
		}
		return nil
	}

	return fmt.Errorf("email service is unavailable")
}
