package appconf

import (
	"log"
	"os"
)

var (
	UserPoolID string
)

func Init() {
	UserPoolID = getEnvOrPanic("COGNITO_USER_POOL_ID")
}

// getEnvOrPanic gets an environment variable or panics if it's not set
func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return value
}
