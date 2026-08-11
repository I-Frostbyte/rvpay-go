package model

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Note: not parallel — tests share the process environment.
	// Clear environment variables to test defaults.
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LISTEN_PORT")
	os.Unsetenv("MIGRATION_PATH")
	os.Unsetenv("RUN_MIGRATIONS")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_TLS_DISABLED")

	cfg := Config{}
	err := cfg.LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig should fail when required variables are missing")
	}
}

func TestLoadConfigRequiredFields(t *testing.T) {
	// Note: not parallel — tests share the process environment.
	// Only LISTEN_PORT provided; MIGRATION_PATH and DB_* are still missing.
	os.Setenv("LISTEN_PORT", "50051")
	defer os.Unsetenv("LISTEN_PORT")

	cfg := Config{}
	err := cfg.LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig should fail when MIGRATION_PATH and DB_* are missing")
	}
}

func TestLoadConfigValidEnvironment(t *testing.T) {
	// Note: not parallel — tests share the process environment.
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LISTEN_PORT", "50051")
	os.Setenv("MIGRATION_PATH", "db/migrations")
	os.Setenv("RUN_MIGRATIONS", "false")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-pass")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_NAME", "transactions")
	os.Setenv("DB_TLS_DISABLED", "true")

	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LISTEN_PORT")
		os.Unsetenv("MIGRATION_PATH")
		os.Unsetenv("RUN_MIGRATIONS")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_TLS_DISABLED")
	}()

	cfg := Config{}
	if err := cfg.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %s, want info", cfg.LogLevel)
	}
	if cfg.ListenPort != "50051" {
		t.Fatalf("ListenPort = %s, want 50051", cfg.ListenPort)
	}
	if cfg.MigrationPath != "db/migrations" {
		t.Fatalf("MigrationPath = %s, want db/migrations", cfg.MigrationPath)
	}
	if cfg.RunMigrations {
		t.Fatal("RunMigrations should be false")
	}
	if cfg.DB.DBUser != "test-user" {
		t.Fatalf("DBUser = %s, want test-user", cfg.DB.DBUser)
	}
	if cfg.DB.DBPassword != "test-pass" {
		t.Fatalf("DBPassword = %s, want test-pass", cfg.DB.DBPassword)
	}
	if cfg.DB.DBHost != "localhost" {
		t.Fatalf("DBHost = %s, want localhost", cfg.DB.DBHost)
	}
	if cfg.DB.DBPort != 5433 {
		t.Fatalf("DBPort = %d, want 5433", cfg.DB.DBPort)
	}
	if cfg.DB.DBName != "transactions" {
		t.Fatalf("DBName = %s, want transactions", cfg.DB.DBName)
	}
	if !cfg.DB.TLSDisabled {
		t.Fatal("TLSDisabled should be true")
	}
}

func TestLoadConfigDefaultsApplied(t *testing.T) {
	// Note: not parallel — tests share the process environment.
	os.Setenv("LISTEN_PORT", "50051")
	os.Setenv("MIGRATION_PATH", "db/migrations")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "transactions")

	defer func() {
		os.Unsetenv("LISTEN_PORT")
		os.Unsetenv("MIGRATION_PATH")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
	}()

	cfg := Config{}
	if err := cfg.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %s, want default debug", cfg.LogLevel)
	}
	if !cfg.RunMigrations {
		t.Fatal("RunMigrations should default to true")
	}
	if cfg.DB.TLSDisabled {
		t.Fatal("TLSDisabled should default to false")
	}
}
