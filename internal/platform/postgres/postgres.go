// Package postgres builds the shared pgx connection pool used by every service.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the connection URL and pool limits.
type Config struct {
	URL               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

// DefaultConfig returns conservative limits for a single-node Postgres.
// Postgres allows 100 connections by default and every service instance
// holds its own pool, so each pool stays small.
func DefaultConfig(url string) Config {
	return Config{
		URL:               url,
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
		ConnectTimeout:    5 * time.Second,
	}
}

// ConfigFromEnv reads DATABASE_URL (required) and optional DB_MAX_CONNS and
// DB_MIN_CONNS overrides.
func ConfigFromEnv() (Config, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return Config{}, errors.New("DATABASE_URL is not set")
	}
	cfg := DefaultConfig(url)
	var err error
	if cfg.MaxConns, err = int32FromEnv("DB_MAX_CONNS", cfg.MaxConns); err != nil {
		return Config{}, err
	}
	if cfg.MinConns, err = int32FromEnv("DB_MIN_CONNS", cfg.MinConns); err != nil {
		return Config{}, err
	}
	return cfg, cfg.Validate()
}

func int32FromEnv(key string, fallback int32) (int32, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return int32(n), nil
}

// Validate rejects configurations that would make the pool misbehave.
func (c Config) Validate() error {
	switch {
	case c.URL == "":
		return errors.New("postgres: URL is empty")
	case c.MaxConns < 1:
		return errors.New("postgres: MaxConns must be at least 1")
	case c.MinConns < 0 || c.MinConns > c.MaxConns:
		return errors.New("postgres: MinConns must be between 0 and MaxConns")
	case c.ConnectTimeout <= 0:
		return errors.New("postgres: ConnectTimeout must be positive")
	}
	return nil
}

// NewPool opens a pool and verifies the database is reachable.
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	pc, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse url: %w", err)
	}
	pc.MaxConns = cfg.MaxConns
	pc.MinConns = cfg.MinConns
	pc.MaxConnLifetime = cfg.MaxConnLifetime
	pc.MaxConnIdleTime = cfg.MaxConnIdleTime
	pc.HealthCheckPeriod = cfg.HealthCheckPeriod
	pc.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return pool, nil
}
