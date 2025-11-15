package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Config captures the configuration required to connect to a SQL database.
type Config struct {
	DSN string
}

// Open creates a database connection. The driver is inferred from the DSN.
func Open(cfg Config) (*sql.DB, error) {
	dsn := cfg.DSN
	if dsn == "" {
		dsn = "file:data/studytracker.db?_pragma=foreign_keys(ON)"
	}

	driver := detectDriver(dsn)
	if driver == "sqlite" {
		if err := ensureDirectory(dsn); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// ApplyMigrations executes embedded SQL migrations in order.
func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("database connection is nil")
	}

	useDollar := UsesDollarPlaceholders(db)

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		var count int
		query := Rebind("SELECT COUNT(1) FROM schema_migrations WHERE name = ?", useDollar)
		if err := db.QueryRowContext(ctx, query, name).Scan(&count); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if count > 0 {
			continue
		}

		sqlBytes, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		insert := Rebind("INSERT INTO schema_migrations (name, applied_at) VALUES (?, ?)", useDollar)
		if _, err := tx.ExecContext(ctx, insert, name, time.Now().UTC()); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}

func ensureDirectory(dsn string) error {
	if strings.HasPrefix(dsn, "file:") {
		path := strings.TrimPrefix(dsn, "file:")
		if idx := strings.Index(path, "?"); idx >= 0 {
			path = path[:idx]
		}
		if path == ":memory:" || path == "" {
			return nil
		}
		return os.MkdirAll(filepath.Dir(path), 0o755)
	}

	if strings.Contains(dsn, "://") {
		// Assume DSN is a URL with driver-specific semantics. Nothing to do.
		return nil
	}

	if dsn == ":memory:" || dsn == "" {
		return nil
	}

	return os.MkdirAll(filepath.Dir(dsn), 0o755)
}

func detectDriver(dsn string) string {
	lower := strings.ToLower(dsn)
	switch {
	case strings.HasPrefix(lower, "postgres://"), strings.HasPrefix(lower, "postgresql://"):
		return "postgres"
	default:
		return "sqlite"
	}
}

// UsesDollarPlaceholders reports whether the active SQL driver expects $1-style placeholders.
func UsesDollarPlaceholders(db *sql.DB) bool {
	driverName := fmt.Sprintf("%T", db.Driver())
	return strings.Contains(driverName, "pq")
}

// Rebind rewrites a query that uses ? placeholders into the driver-specific format.
func Rebind(query string, useDollar bool) string {
	if !useDollar {
		return query
	}
	var (
		builder strings.Builder
		index   = 1
	)
	builder.Grow(len(query) + 10)
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			builder.WriteString(fmt.Sprintf("$%d", index))
			index++
			continue
		}
		builder.WriteByte(query[i])
	}
	return builder.String()
}
