package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Runtime struct {
	Environment     string
	HTTPAddress     string
	DatabaseURL     string
	StorageRoot     string
	Timezone        string
	AllowedOrigins  []string
	JWTIssuer       string
	JWTSigningKey   string
	ShutdownTimeout time.Duration
	RequestTimeout  time.Duration
	WorkerInterval  time.Duration
}

func FromEnvironment() (Runtime, error) {
	runtime := Runtime{
		Environment:     env("APP_ENV", "development"),
		HTTPAddress:     env("HTTP_ADDRESS", ":8080"),
		DatabaseURL:     env("DATABASE_URL", "postgres://cry092:cry092@localhost:5432/cry092?sslmode=disable"),
		StorageRoot:     env("FILE_STORAGE_ROOT", "./var/files"),
		Timezone:        env("APP_TIMEZONE", "Asia/Shanghai"),
		AllowedOrigins:  split(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		JWTIssuer:       env("JWT_ISSUER", "cry-092"),
		JWTSigningKey:   os.Getenv("JWT_SIGNING_KEY"),
		ShutdownTimeout: duration("SHUTDOWN_TIMEOUT", 15*time.Second),
		RequestTimeout:  duration("REQUEST_TIMEOUT", 8*time.Second),
		WorkerInterval:  duration("WORKER_INTERVAL", 30*time.Second),
	}
	if err := runtime.Check(); err != nil {
		return Runtime{}, err
	}
	return runtime, nil
}

func (c Runtime) Check() error {
	if c.HTTPAddress == "" || c.DatabaseURL == "" || c.StorageRoot == "" {
		return fmt.Errorf("http address, database URL and storage root are required")
	}
	if len(c.JWTSigningKey) < 32 {
		return fmt.Errorf("JWT_SIGNING_KEY must contain at least 32 characters")
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return fmt.Errorf("invalid APP_TIMEZONE: %w", err)
	}
	if c.RequestTimeout <= 0 || c.ShutdownTimeout <= 0 || c.WorkerInterval <= 0 {
		return fmt.Errorf("timeouts and worker interval must be positive")
	}
	return nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func split(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func duration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
