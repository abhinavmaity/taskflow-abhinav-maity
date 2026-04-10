package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	JWTSecret   string
	JWTIssuer   string
	JWTTTL      time.Duration
	LogLevel    slog.Level
}

func Load() (Config, error) {
	cfg := Config{
		ServerPort:  getOrDefault("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTIssuer:   getOrDefault("JWT_ISSUER", "taskflow"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	if err := validatePort(cfg.ServerPort); err != nil {
		return Config{}, err
	}

	ttlHours := getOrDefault("JWT_TTL_HOURS", "24")
	hours, err := strconv.Atoi(ttlHours)
	if err != nil || hours <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL_HOURS must be a positive integer")
	}
	cfg.JWTTTL = time.Duration(hours) * time.Hour

	cfg.LogLevel = parseLevel(getOrDefault("LOG_LEVEL", "info"))

	return cfg, nil
}

func getOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n <= 0 || n > 65535 {
		return fmt.Errorf("PORT must be a valid TCP port")
	}
	return nil
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
