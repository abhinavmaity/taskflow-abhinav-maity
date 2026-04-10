package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/bootstrap"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	seedDir := strings.TrimSpace(os.Getenv("SEED_DIR"))
	if seedDir == "" {
		seedDir = "./seed"
	}

	enabled := parseEnabled(strings.TrimSpace(os.Getenv("SEED_ENABLED")))

	ctx := context.Background()
	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := bootstrap.RunSeed(ctx, pool, seedDir, enabled, logger); err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}

	logger.Info("seed command completed")
}

func parseEnabled(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
