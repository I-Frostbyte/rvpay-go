package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("TOKEN_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("LISTEN_PORT", "50051")
	t.Setenv("MIGRATION_PATH", "db/migrations")
	t.Setenv("DB_USER", "rvpay")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "rvpay")

	config := Config{}
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.LogLevel != "debug" {
		t.Fatalf("log_level = %s, want debug (default)", config.LogLevel)
	}
	if !config.RunMigrations {
		t.Fatal("run_migrations = false, want true (default)")
	}
	if config.DB.TLSDisabled {
		t.Fatal("db_tls_disabled = true, want false (default)")
	}
	if config.ListenPort != "50051" {
		t.Fatalf("listen_port = %s, want 50051", config.ListenPort)
	}
	if config.MigrationPath != "db/migrations" {
		t.Fatalf("migration_path = %s, want db/migrations", config.MigrationPath)
	}
	if config.TokenEncryptionKey != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("token_encryption_key = %s, want test key", config.TokenEncryptionKey)
	}
}

func TestLoadConfigOverridesDefaults(t *testing.T) {
	t.Setenv("TOKEN_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("LISTEN_PORT", "50051")
	t.Setenv("MIGRATION_PATH", "db/migrations")
	t.Setenv("DB_USER", "rvpay")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "rvpay")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("RUN_MIGRATIONS", "false")
	t.Setenv("DB_TLS_DISABLED", "true")
	t.Setenv("env", "production")

	config := Config{}
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.LogLevel != "info" {
		t.Fatalf("log_level = %s, want info", config.LogLevel)
	}
	if config.RunMigrations {
		t.Fatal("run_migrations = true, want false")
	}
	if !config.DB.TLSDisabled {
		t.Fatal("db_tls_disabled = false, want true")
	}
	if config.Env != "production" {
		t.Fatalf("env = %s, want production", config.Env)
	}
}

func TestLoadConfigMissingRequiredVariables(t *testing.T) {
	tests := []struct {
		name    string
		unset   []string
		envVars map[string]string
	}{
		{
			name:  "missing token encryption key",
			unset: []string{"TOKEN_ENCRYPTION_KEY"},
		},
		{
			name:  "missing listen port",
			unset: []string{"LISTEN_PORT"},
		},
		{
			name:  "missing migration path",
			unset: []string{"MIGRATION_PATH"},
		},
		{
			name:  "missing db user",
			unset: []string{"DB_USER"},
		},
		{
			name:  "missing db password",
			unset: []string{"DB_PASSWORD"},
		},
		{
			name:  "missing db host",
			unset: []string{"DB_HOST"},
		},
		{
			name:  "missing db port",
			unset: []string{"DB_PORT"},
		},
		{
			name:  "missing db name",
			unset: []string{"DB_NAME"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set all required variables first.
			t.Setenv("TOKEN_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
			t.Setenv("LISTEN_PORT", "50051")
			t.Setenv("MIGRATION_PATH", "db/migrations")
			t.Setenv("DB_USER", "rvpay")
			t.Setenv("DB_PASSWORD", "secret")
			t.Setenv("DB_HOST", "localhost")
			t.Setenv("DB_PORT", "5432")
			t.Setenv("DB_NAME", "rvpay")

			// Unset the variable under test.
			for _, key := range tt.unset {
				original, existed := os.LookupEnv(key)
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("failed to unset %s: %v", key, err)
				}
				if existed {
					t.Cleanup(func() {
						_ = os.Setenv(key, original)
					})
				}
			}

			config := Config{}
			if err := config.LoadConfig(); err == nil {
				t.Fatalf("expected error for missing %v, got nil", tt.unset)
			}
		})
	}
}

func TestLoadConfigFromEnvFile(t *testing.T) {
	// Create a temporary directory with a .env file and chdir into it so
	// LoadConfig picks up the file from the working directory.
	dir := t.TempDir()
	envContent := `TOKEN_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef
LISTEN_PORT=60000
MIGRATION_PATH=db/migrations
DB_USER=envfile-user
DB_PASSWORD=envfile-pass
DB_HOST=envfile-host
DB_PORT=5433
DB_NAME=envfile-db
LOG_LEVEL=warn
`
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0o600); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	t.Chdir(dir)

	config := Config{}
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.TokenEncryptionKey != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("token_encryption_key = %s, want env file value", config.TokenEncryptionKey)
	}
	if config.ListenPort != "60000" {
		t.Fatalf("listen_port = %s, want 60000", config.ListenPort)
	}
	if config.DB.DBUser != "envfile-user" {
		t.Fatalf("db_user = %s, want envfile-user", config.DB.DBUser)
	}
	if config.DB.DBPassword != "envfile-pass" {
		t.Fatalf("db_password = %s, want envfile-pass", config.DB.DBPassword)
	}
	if config.DB.DBHost != "envfile-host" {
		t.Fatalf("db_host = %s, want envfile-host", config.DB.DBHost)
	}
	if config.DB.DBPort != 5433 {
		t.Fatalf("db_port = %d, want 5433", config.DB.DBPort)
	}
	if config.DB.DBName != "envfile-db" {
		t.Fatalf("db_name = %s, want envfile-db", config.DB.DBName)
	}
	if config.LogLevel != "warn" {
		t.Fatalf("log_level = %s, want warn", config.LogLevel)
	}
}

func TestLoadConfigInvalidEnvFile(t *testing.T) {
	// A malformed .env file should cause LoadConfig to return an error.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("THIS_IS_NOT_VALID_ENV_SYNTAX\n"), 0o600); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	t.Chdir(dir)

	config := Config{}
	err := config.LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid .env file, got nil")
	}
	if !strings.Contains(err.Error(), "error loading .env file") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestLoadConfigHighLevelVariables(t *testing.T) {
	t.Setenv("TOKEN_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("LISTEN_PORT", "50051")
	t.Setenv("MIGRATION_PATH", "db/migrations")
	t.Setenv("DB_USER", "rvpay")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "rvpay")
	t.Setenv("HIGHLEVEL_CLIENT_ID", "client-id")
	t.Setenv("HIGHLEVEL_CLIENT_SECRET", "client-secret")
	t.Setenv("HIGHLEVEL_REDIRECT_URL", "http://localhost:8080/oauth/callback")
	t.Setenv("HIGHLEVEL_SSO_KEY", "sso-key")

	config := Config{}
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.HighLevelClientID != "client-id" {
		t.Fatalf("highlevel_client_id = %s, want client-id", config.HighLevelClientID)
	}
	if config.HighLevelClientSecret != "client-secret" {
		t.Fatalf("highlevel_client_secret = %s, want client-secret", config.HighLevelClientSecret)
	}
	if config.HighLevelRedirectURL != "http://localhost:8080/oauth/callback" {
		t.Fatalf("highlevel_redirect_url = %s, want http://localhost:8080/oauth/callback", config.HighLevelRedirectURL)
	}
	if config.HighLevelSSOKey != "sso-key" {
		t.Fatalf("highlevel_sso_key = %s, want sso-key", config.HighLevelSSOKey)
	}
}
