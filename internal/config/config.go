package config

import (
	"fmt"
	"os"
	"path/filepath"
	"payment-service/internal/database"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds application settings loaded from YAML (per APP_ENV), then overridden by environment.
type Config struct {
	HTTPPort                  string                          `koanf:"http_port"`
	Database                  database.DatabaseConfig         `koanf:"database"`
	PaymentGateway            string                          `koanf:"payment_gateway"`
	PaymentGatewayPriority    []string                        `koanf:"payment_gateway_priority"`
	PaymentCommunicationRetry PaymentCommunicationRetryConfig `koanf:"payment_communication_retry"`
}

// PaymentCommunicationRetryConfig drives the background poller that re-calls the gateway for stuck communications.
type PaymentCommunicationRetryConfig struct {
	Enabled                   bool     `koanf:"enabled"`
	IntervalSeconds           int      `koanf:"interval_seconds"`
	RetryableResponseStatuses []string `koanf:"retryable_response_statuses"`
	MaxPollAttempts           int      `koanf:"max_poll_attempts"`
	BatchLimit                int      `koanf:"batch_limit"`
}

// Load reads configs/<config.{dev|prod}.yaml> (see APP_ENV), merges APP_* env vars, then
// standard names PORT, DATABASE_URL, and PAYMENT_GATEWAY.
func Load() (Config, error) {
	k := koanf.New(".")

	suffix, err := resolveEnvSuffix()
	if err != nil {
		return Config{}, err
	}

	configDir := os.Getenv("CONFIG_DIR")

	if configDir == "" {
		configDir = "configs"
	}
	path := filepath.Join(configDir, fmt.Sprintf("config.%s.yaml", suffix))

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return Config{}, fmt.Errorf("load %s: %w", path, err)
	}

	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: "APP_",
		TransformFunc: func(envKey, v string) (string, any) {
			key := strings.ToLower(strings.TrimPrefix(envKey, "APP_"))
			return key, v
		},
	}), nil); err != nil {
		return Config{}, fmt.Errorf("merge APP_ env: %w", err)
	}

	if p := os.Getenv("PORT"); p != "" {
		if err := k.Set("http_port", p); err != nil {
			return Config{}, err
		}
	}

	if d := os.Getenv("DATABASE_URL"); d != "" {
		if err := k.Set("database_url", d); err != nil {
			return Config{}, err
		}
	}

	if g := os.Getenv("PAYMENT_GATEWAY"); g != "" {
		if err := k.Set("payment_gateway", g); err != nil {
			return Config{}, err
		}
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8080"
	}

	if cfg.Database.Dsn == "" {
		return Config{}, fmt.Errorf("database_url is required (configs YAML or DATABASE_URL / APP_DATABASE_URL)")
	}

	if cfg.PaymentGateway == "" {
		cfg.PaymentGateway = "stub"
	}

	applyPaymentCommunicationRetryDefaults(&cfg.PaymentCommunicationRetry)

	return cfg, nil
}

func resolveEnvSuffix() (string, error) {
	envName := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if envName == "" {
		return "dev", nil
	}
	if envName == "production" {
		return "prod", nil
	}
	switch envName {
	case "dev", "prod":
		return envName, nil
	default:
		return "", fmt.Errorf("APP_ENV must be dev, prod, or production; got %q", envName)
	}
}

func applyPaymentCommunicationRetryDefaults(c *PaymentCommunicationRetryConfig) {
	if !c.Enabled {
		return
	}
	if c.IntervalSeconds <= 0 {
		c.IntervalSeconds = 10
	}
	if len(c.RetryableResponseStatuses) == 0 {
		c.RetryableResponseStatuses = []string{"pending"}
	}
	if c.MaxPollAttempts <= 0 {
		c.MaxPollAttempts = 30
	}
	if c.BatchLimit <= 0 {
		c.BatchLimit = 100
	}
}
