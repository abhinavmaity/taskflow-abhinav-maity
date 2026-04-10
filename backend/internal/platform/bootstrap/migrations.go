package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string, logger *slog.Logger) error {
	if err := ensureMigrationsTable(ctx, pool); err != nil {
		return err
	}

	applied, err := loadAppliedMigrations(ctx, pool)
	if err != nil {
		return err
	}

	files, err := sqlFiles(migrationsDir, ".up.sql")
	if err != nil {
		return err
	}

	for _, migration := range files {
		version := strings.TrimSuffix(filepath.Base(migration), ".up.sql")
		if _, ok := applied[version]; ok {
			continue
		}

		content, err := os.ReadFile(migration)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", migration, err)
		}

		if _, err := pool.Exec(ctx, string(content), pgx.QueryExecModeSimpleProtocol); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration, err)
		}

		if _, err := pool.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}

		logger.Info("migration applied", "version", version, "file", filepath.Base(migration))
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`
	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func loadAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[string]struct{}, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	out := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		out[version] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return out, nil
}

func sqlFiles(dir, suffix string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read sql directory %s: %w", dir, err)
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}

	slices.Sort(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no migration files found in %s", dir)
	}
	return files, nil
}
