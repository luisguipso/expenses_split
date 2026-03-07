package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	FrontendURL         string
	SMTPHost            string
	SMTPPort            int
	SMTPUsername        string
	SMTPPassword        string
	SMTPFrom            string
	VerificationCodeTTL time.Duration
	PasswordResetTTL    time.Duration
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://contas:contas@localhost:5432/contas?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:5173"),
		SMTPHost:            getEnv("SMTP_HOST", "localhost"),
		SMTPPort:            getEnvInt("SMTP_PORT", 587),
		SMTPUsername:        getEnv("SMTP_USERNAME", ""),
		SMTPPassword:        getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:            getEnv("SMTP_FROM", "noreply@contas.app"),
		VerificationCodeTTL: getEnvDuration("VERIFICATION_CODE_TTL", 15*time.Minute),
		PasswordResetTTL:    getEnvDuration("PASSWORD_RESET_TTL", 30*time.Minute),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
