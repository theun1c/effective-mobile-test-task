package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsDotEnvFile(t *testing.T) {
	unsetEnv(t, "APP_ENV", "LOG_LEVEL", "HTTP_HOST", "HTTP_PORT", "HTTP_SHUTDOWN_TIMEOUT", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSLMODE", "DB_PING_TIMEOUT")

	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	content := []byte("APP_ENV=test\nHTTP_PORT=18080\nDB_HOST=postgres\nDB_PASSWORD=secret\n")
	if err := os.WriteFile(envFile, content, 0o644); err != nil {
		t.Fatalf("write .env file: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	t.Cleanup(func() {
		if chdirErr := os.Chdir(previousDir); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.AppEnv != "test" {
		t.Fatalf("expected APP_ENV from .env, got %q", cfg.AppEnv)
	}

	if cfg.HTTP.Port != "18080" {
		t.Fatalf("expected HTTP_PORT from .env, got %q", cfg.HTTP.Port)
	}

	if cfg.Postgres.Host != "postgres" {
		t.Fatalf("expected DB_HOST from .env, got %q", cfg.Postgres.Host)
	}

	if cfg.Postgres.Password != "secret" {
		t.Fatalf("expected DB_PASSWORD from .env, got %q", cfg.Postgres.Password)
	}
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()

	previous := make(map[string]*string, len(keys))
	for _, key := range keys {
		value, exists := os.LookupEnv(key)
		if exists {
			valueCopy := value
			previous[key] = &valueCopy
		} else {
			previous[key] = nil
		}

		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for _, key := range keys {
			value := previous[key]
			var err error
			if value == nil {
				err = os.Unsetenv(key)
			} else {
				err = os.Setenv(key, *value)
			}

			if err != nil {
				t.Fatalf("restore %s: %v", key, err)
			}
		}
	})
}
