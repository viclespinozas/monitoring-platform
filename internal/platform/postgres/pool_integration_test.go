//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/viclespinozas/monitoring-platform/internal/platform/postgres"
	"github.com/viclespinozas/monitoring-platform/internal/testutil/postgrestest"
)

var testDB *postgrestest.DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	db, err := postgrestest.Start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot start postgres (is Docker running?):", err)
		os.Exit(1)
	}
	testDB = db
	code := m.Run()
	db.Stop(ctx)
	os.Exit(code)
}

func TestNewPoolConnectsAndAppliesLimits(t *testing.T) {
	cfg := postgres.DefaultConfig(testDB.URL)
	cfg.MaxConns = 3
	pool, err := postgres.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if got := pool.Config().MaxConns; got != 3 {
		t.Fatalf("MaxConns = %d, want 3", got)
	}
	var one int
	if err := pool.QueryRow(context.Background(), "SELECT 1").Scan(&one); err != nil || one != 1 {
		t.Fatalf("query through pool failed: %v (got %d)", err, one)
	}
}

func TestNewPoolFailsFastWhenUnreachable(t *testing.T) {
	cfg := postgres.DefaultConfig("postgres://test:test@127.0.0.1:1/test?sslmode=disable")
	cfg.ConnectTimeout = 500 * time.Millisecond

	start := time.Now()
	if _, err := postgres.NewPool(context.Background(), cfg); err == nil {
		t.Fatal("expected an error for an unreachable database")
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("took %v to fail, ConnectTimeout is not respected", elapsed)
	}
}
