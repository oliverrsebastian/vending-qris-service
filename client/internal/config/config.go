package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds client runtime settings. Secrets are injected by the environment (CI/CD), never hard-coded.
type Config struct {
	BaseURL string
	AuthKey string
}

func Load() (Config, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PAYMENT_SERVICE_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	authKey := strings.TrimSpace(os.Getenv("PAYMENT_SERVICE_AUTH_KEY"))

	return Config{
		BaseURL: baseURL,
		AuthKey: authKey,
	}, nil
}

func (c Config) RequireAuthKey() error {
	if c.AuthKey == "" {
		return fmt.Errorf("PAYMENT_SERVICE_AUTH_KEY is required for admin endpoints")
	}
	return nil
}
