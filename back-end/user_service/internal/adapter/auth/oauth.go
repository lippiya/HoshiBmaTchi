package auth

import (
	"context"
	"hoshiBmaTchi/back-end/user_service/internal/core/domain"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var config = &oauth2.Config{
    ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
    Scopes: []string{
        "https://www.googleapis.com/auth/userinfo.email",
        "https://www.googleapis.com/auth/userinfo.profile",
    },
    Endpoint: google.Endpoint,
}


func GetGoogleUserInfo(code string) (*domain.OAuthUserData, error) {
    token, err := config.Exchange(context.Background(), code)
    if err != nil {
        return nil, err
    }
    
    client := config.Client(context.Background(), token)
    oauth2Service, err := oauth2api.NewService(context.Background(), option.WithHTTPClient(client))
    if err != nil {
        return nil, err
    }
    
    userInfo, err := oauth2Service.Userinfo.Get().Do()
    if err != nil {
        return nil, err
    }
    
    return &domain.OAuthUserData{
        ID: userInfo.Id,
        Email: userInfo.Email,
        VerifiedEmail: userInfo.VerifiedEmail != nil && *userInfo.VerifiedEmail,
        Name: userInfo.Name,
        Picture: userInfo.Picture,
    }, nil
}

func GenerateAuthURL(state string) string {
    return config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}