package service

import (
	"hoshiBmaTchi/back-end/user_service/internal/adapter/auth"
	"hoshiBmaTchi/back-end/user_service/internal/core/domain"
)

type OAuthService struct{}

func NewOAuthService() *OAuthService {
	return &OAuthService{}
}

func (o *OAuthService) GetGoogleUserInfo(code string) (*domain.OAuthUserData, error) {
	return auth.GetGoogleUserInfo(code)
}

func (o *OAuthService) GenerateAuthURL(state string) string {
	return auth.GenerateAuthURL(state)
}
