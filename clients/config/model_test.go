package config

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Parallel()

	// Clear environment variables to test defaults
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LISTEN_PORT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_TLS_DISABLED")
	os.Unsetenv("RUN_MIGRATIONS")
	os.Unsetenv("MIGRATION_PATH")
	os.Unsetenv("HIGHLEVEL_CLIENT_ID")
	os.Unsetenv("HIGHLEVEL_CLIENT_SECRET")
	os.Unsetenv("HIGHLEVEL_REDIRECT_URI")
	os.Unsetenv("WEBHOOK_SECRET")

	cfg := Config{}
	err := cfg.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %s, want info", cfg.LogLevel)
	}
	if cfg.ListenPort != 50051 {
		t.Fatalf("ListenPort = %d, want 50051", cfg.ListenPort)
	}
	if cfg.DB.DBHost != "localhost" {
		t.Fatalf("DBHost = %s, want localhost", cfg.DB.DBHost)
	}
	if cfg.DB.DBPort != 5432 {
		t.Fatalf("DBPort = %d, want 5432", cfg.DB.DBPort)
	}
	if cfg.DB.DBName != "rvpay" {
		t.Fatalf("DBName = %s, want rvpay", cfg.DB.DBName)
	}
	if cfg.DB.DBUser != "postgres" {
		t.Fatalf("DBUser = %s, want postgres", cfg.DB.DBUser)
	}
	if !cfg.DB.TLSDisabled {
		t.Fatal("TLSDisabled should default to true")
	}
	if !cfg.RunMigrations {
		t.Fatal("RunMigrations should default to true")
	}
	if cfg.MigrationPath != "db/migrations" {
		t.Fatalf("MigrationPath = %s, want db/migrations", cfg.MigrationPath)
	}
	if cfg.HighLevel.RedirectURI != "https://api.rvpay.com/v1/public/oauth/callback" {
		t.Fatalf("RedirectURI = %s, want default", cfg.HighLevel.RedirectURI)
	}
}

func TestLoadConfigEnvironmentOverrides(t *testing.T) {
	t.Parallel()

	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LISTEN_PORT", "60000")
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_NAME", "test-db")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-pass")
	os.Setenv("DB_TLS_DISABLED", "false")
	os.Setenv("RUN_MIGRATIONS", "false")
	os.Setenv("MIGRATION_PATH", "test/migrations")
	os.Setenv("HIGHLEVEL_CLIENT_ID", "test-client-id")
	os.Setenv("HIGHLEVEL_CLIENT_SECRET", "test-client-secret")
	os.Setenv("HIGHLEVEL_REDIRECT_URI", "https://test.example.com/callback")
	os.Setenv("WEBHOOK_SECRET", "test-webhook-secret")

	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LISTEN_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_TLS_DISABLED")
		os.Unsetenv("RUN_MIGRATIONS")
		os.Unsetenv("MIGRATION_PATH")
		os.Unsetenv("HIGHLEVEL_CLIENT_ID")
		os.Unsetenv("HIGHLEVEL_CLIENT_SECRET")
		os.Unsetenv("HIGHLEVEL_REDIRECT_URI")
		os.Unsetenv("WEBHOOK_SECRET")
	}()

	cfg := Config{}
	err := cfg.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %s, want debug", cfg.LogLevel)
	}
	if cfg.ListenPort != 60000 {
		t.Fatalf("ListenPort = %d, want 60000", cfg.ListenPort)
	}
	if cfg.DB.DBHost != "test-host" {
		t.Fatalf("DBHost = %s, want test-host", cfg.DB.DBHost)
	}
	if cfg.DB.DBPort != 5433 {
		t.Fatalf("DBPort = %d, want 5433", cfg.DB.DBPort)
	}
	if cfg.DB.DBName != "test-db" {
		t.Fatalf("DBName = %s, want test-db", cfg.DB.DBName)
	}
	if cfg.DB.DBUser != "test-user" {
		t.Fatalf("DBUser = %s, want test-user", cfg.DB.DBUser)
	}
	if cfg.DB.DBPassword != "test-pass" {
		t.Fatalf("DBPassword = %s, want test-pass", cfg.DB.DBPassword)
	}
	if cfg.DB.TLSDisabled {
		t.Fatal("TLSDisabled should be false")
	}
	if cfg.RunMigrations {
		t.Fatal("RunMigrations should be false")
	}
	if cfg.MigrationPath != "test/migrations" {
		t.Fatalf("MigrationPath = %s, want test/migrations", cfg.MigrationPath)
	}
	if cfg.HighLevel.ClientID != "test-client-id" {
		t.Fatalf("ClientID = %s, want test-client-id", cfg.HighLevel.ClientID)
	}
	if cfg.HighLevel.ClientSecret != "test-client-secret" {
		t.Fatalf("ClientSecret = %s, want test-client-secret", cfg.HighLevel.ClientSecret)
	}
	if cfg.HighLevel.RedirectURI != "https://test.example.com/callback" {
		t.Fatalf("RedirectURI = %s, want test URL", cfg.HighLevel.RedirectURI)
	}
	if cfg.Webhook.Secret != "test-webhook-secret" {
		t.Fatalf("WebhookSecret = %s, want test-webhook-secret", cfg.Webhook.Secret)
	}
}

func TestLoadConfigInvalidValues(t *testing.T) {
	t.Parallel()

	os.Setenv("LISTEN_PORT", "invalid")
	os.Setenv("DB_PORT", "invalid")
	os.Setenv("DB_TLS_DISABLED", "invalid")
	os.Setenv("RUN_MIGRATIONS", "invalid")

	defer func() {
		os.Unsetenv("LISTEN_PORT")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_TLS_DISABLED")
		os.Unsetenv("RUN_MIGRATIONS")
	}()

	cfg := Config{}
	err := cfg.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Invalid values should fall back to defaults
	if cfg.ListenPort != 50051 {
		t.Fatalf("ListenPort = %d, want default 50051", cfg.ListenPort)
	}
	if cfg.DB.DBPort != 5432 {
		t.Fatalf("DBPort = %d, want default 5432", cfg.DB.DBPort)
	}
	if !cfg.DB.TLSDisabled {
		t.Fatal("TLSDisabled should default to true for invalid value")
	}
	if !cfg.RunMigrations {
		t.Fatal("RunMigrations should default to true for invalid value")
	}
}