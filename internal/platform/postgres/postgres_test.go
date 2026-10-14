package postgres

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	valid := DefaultConfig("postgres://u:p@localhost:5432/db")

	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"defaults are valid", func(*Config) {}, false},
		{"empty url", func(c *Config) { c.URL = "" }, true},
		{"zero max conns", func(c *Config) { c.MaxConns = 0 }, true},
		{"min above max", func(c *Config) { c.MinConns = c.MaxConns + 1 }, true},
		{"negative min", func(c *Config) { c.MinConns = -1 }, true},
		{"zero min is fine", func(c *Config) { c.MinConns = 0 }, false},
		{"zero connect timeout", func(c *Config) { c.ConnectTimeout = 0 }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)
			if err := cfg.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Run("missing url", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		if _, err := ConfigFromEnv(); err == nil {
			t.Fatal("expected error when DATABASE_URL is unset")
		}
	})

	t.Run("defaults", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("DB_MAX_CONNS", "")
		t.Setenv("DB_MIN_CONNS", "")
		cfg, err := ConfigFromEnv()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.MaxConns != 10 || cfg.MinConns != 2 || cfg.ConnectTimeout != 5*time.Second {
			t.Fatalf("unexpected defaults: %+v", cfg)
		}
	})

	t.Run("overrides", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("DB_MAX_CONNS", "20")
		t.Setenv("DB_MIN_CONNS", "5")
		cfg, err := ConfigFromEnv()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.MaxConns != 20 || cfg.MinConns != 5 {
			t.Fatalf("overrides not applied: %+v", cfg)
		}
	})

	t.Run("not a number", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("DB_MAX_CONNS", "lots")
		if _, err := ConfigFromEnv(); err == nil {
			t.Fatal("expected error for non-numeric DB_MAX_CONNS")
		}
	})
}
