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

	migrationsDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := bootstrap.RunMigrations(ctx, pool, migrationsDir, logger); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations completed")
}
