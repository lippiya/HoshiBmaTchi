package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type SpamService struct{}

func NewSpamService() *SpamService {
	return &SpamService{}
}

type TurnstileResponse struct {
	Success bool `json:"success"`
}

func (s *SpamService) Validate(token string) error {
	secret := os.Getenv("TURNSTILE_API_SECRET")

	if token == "" {
		return errors.New("turnstile token is empty")
	}
	if secret == "" {
		return errors.New("TURNSTILE_API_SECRET not configured")
	}

	response, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify",
		url.Values{
			"secret":   {secret},
			"response": {token},
		},
	)
	if err != nil {
		log.Printf("Turnstile HTTP request failed: %v", err)
		return errors.New("failed to verify captcha")
	}
	defer response.Body.Close()

	var tr TurnstileResponse
	err = json.NewDecoder(response.Body).Decode(&tr)
	if err != nil {
		log.Printf("Failed to decode Turnstile response: %v", err)
		return errors.New("failed to parse captcha response")
	}

	if !tr.Success {
		return fmt.Errorf("captcha verification failed - Turnstile returned success=false")
	}
	return nil
}
