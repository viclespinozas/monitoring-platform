//go:build integration

// Package postgrestest starts a throwaway Postgres container for integration
// tests and applies the project's migrations to it.
//
// Typical use is one container per test package, started in TestMain:
//
//	func TestMain(m *testing.M) {
//		db, err := postgrestest.Start(context.Background())
//		...
//		defer db.Stop(context.Background())
//		os.Exit(m.Run())
//	}
//
// and db.Reset(t) at the top of each test for a clean services table.
package postgrestest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const image = "postgres:17-alpine"

// DB is a running Postgres container with a connected pool.
type DB struct {
	// URL is the connection string for the container.
	URL  string
	Pool *pgxpool.Pool

	container testcontainers.Container
}

// Start launches Postgres, applies every migration and opens a pool.
func Start(ctx context.Context) (*DB, error) {
	c, err := tcpostgres.Run(ctx, image,
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	url, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(c)
		return nil, fmt.Errorf("connection string: %w", err)
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		_ = testcontainers.TerminateContainer(c)
		return nil, fmt.Errorf("open pool: %w", err)
	}

	db := &DB{URL: url, Pool: pool, container: c}
	if err := db.migrate(ctx); err != nil {
		db.Stop(ctx)
		return nil, err
	}
	return db, nil
}

// Stop closes the pool and terminates the container.
func (d *DB) Stop(_ context.Context) {
	d.Pool.Close()
	_ = testcontainers.TerminateContainer(d.container)
}

// Reset empties the tables so each test starts from a known state.
func (d *DB) Reset(t testing.TB) {
	t.Helper()
	if _, err := d.Pool.Exec(context.Background(), `TRUNCATE services`); err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

// migrate runs the "-- +goose Up" section of every migration file in order.
// goose itself is not a module dependency (see ADR-0005), and the Up sections
// are plain SQL, so applying them directly keeps the schema identical to what
// `make migrate-up` produces.
func (d *DB) migrate(ctx context.Context) error {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "migrations")

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil || len(files) == 0 {
		return fmt.Errorf("no migrations found in %s (err %v)", dir, err)
	}
	sort.Strings(files)

	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := d.Pool.Exec(ctx, upSection(string(raw))); err != nil {
			return fmt.Errorf("apply %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

func upSection(sql string) string {
	_, after, _ := strings.Cut(sql, "-- +goose Up")
	up, _, _ := strings.Cut(after, "-- +goose Down")
	return up
}
