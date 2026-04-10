package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RunSeed(ctx context.Context, pool *pgxpool.Pool, seedDir string, enabled bool, logger *slog.Logger) error {
	if !enabled {
		logger.Info("seed skipped", "reason", "SEED_ENABLED=false")
		return nil
	}

	seedFile := filepath.Join(seedDir, "001_seed.sql")
	content, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("read seed file %s: %w", seedFile, err)
	}

	if _, err := pool.Exec(ctx, string(content), pgx.QueryExecModeSimpleProtocol); err != nil {
		return fmt.Errorf("execute seed file %s: %w", seedFile, err)
	}

	logger.Info("seed applied", "file", filepath.Base(seedFile))
	return nil
}
