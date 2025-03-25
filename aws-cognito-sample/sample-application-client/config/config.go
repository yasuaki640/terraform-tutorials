package config

import (
	"context"
	"log"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	Provider     *oidc.Provider
	Oauth2Config oauth2.Config
)

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	clientID := getEnvOrPanic("CLIENT_ID")
	clientSecret := getEnvOrPanic("CLIENT_SECRET")
	redirectURL := getEnvOrPanic("REDIRECT_URL")
	issuerURL := getEnvOrPanic("ISSUER_URL")

	var err error
	Provider, err = oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		log.Fatalf("Failed to create OIDC provider: %v", err)
	}

	// Set up OAuth2 config
	Oauth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     Provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "openid", "profile"},
	}
}

// getEnvOrPanic gets an environment variable or panics if it's not set
func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return value
}
