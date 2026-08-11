package database

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestPostgresURL_TLSDisabled(t *testing.T) {
	got := PostgresURL("postgres", "secret", 5432, "localhost", "rvpay", true)

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("PostgresURL returned unparseable URL %q: %v", got, err)
	}

	if u.Scheme != "postgres" {
		t.Errorf("scheme = %q, want %q", u.Scheme, "postgres")
	}
	if u.User.Username() != "postgres" {
		t.Errorf("user = %q, want %q", u.User.Username(), "postgres")
	}
	if pass, _ := u.User.Password(); pass != "secret" {
		t.Errorf("password = %q, want %q", pass, "secret")
	}
	if u.Host != "localhost:5432" {
		t.Errorf("host = %q, want %q", u.Host, "localhost:5432")
	}
	if u.Path != "/rvpay" {
		t.Errorf("path = %q, want %q", u.Path, "/rvpay")
	}
	if got := u.Query().Get("sslmode"); got != "disable" {
		t.Errorf("sslmode = %q, want %q", got, "disable")
	}
}

func TestPostgresURL_TLSEnabled(t *testing.T) {
	got := PostgresURL("postgres", "secret", 5433, "db.internal", "deposits", false)

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("PostgresURL returned unparseable URL %q: %v", got, err)
	}

	if u.Host != "db.internal:5433" {
		t.Errorf("host = %q, want %q", u.Host, "db.internal:5433")
	}
	if got := u.Query().Get("sslmode"); got != "require" {
		t.Errorf("sslmode = %q, want %q", got, "require")
	}
}

func TestPostgresURL_PasswordIsEscaped(t *testing.T) {
	// A password containing reserved characters must be URL-encoded so the
	// DSN remains parseable.
	got := PostgresURL("postgres", "p@ss:w/o rd", 5432, "localhost", "rvpay", true)

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("PostgresURL returned unparseable URL %q: %v", got, err)
	}

	pass, _ := u.User.Password()
	if pass != "p@ss:w/o rd" {
		t.Errorf("decoded password = %q, want %q", pass, "p@ss:w/o rd")
	}
}

func TestConnect_InvalidURL(t *testing.T) {
	// An unparseable DSN must fail at pool construction without touching a
	// network, keeping the test deterministic and infrastructure-free.
	_, err := Connect(context.Background(), "::not-a-url::")
	if err == nil {
		t.Fatal("Connect(invalid URL) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to connect to database") {
		t.Errorf("error %q missing context", err)
	}
}

func TestMigrate_MissingMigrationPath(t *testing.T) {
	// A non-existent migration directory must fail deterministically at
	// migrate.New without a database.
	err := Migrate("postgres://postgres:secret@localhost:5432/rvpay?sslmode=disable", "/path/that/does/not/exist", zerolog.Nop())
	if err == nil {
		t.Fatal("Migrate(missing path) expected error, got nil")
	}
}

func TestMigrate_DoesNotLogDatabaseCredentials(t *testing.T) {
	// SECURITY REGRESSION TEST (SEC-01): the migrate.Migrate struct holds the
	// full database URL including credentials. It must never be written to
	// logs. This test captures the logger output and verifies the password and
	// the full DSN do not appear.
	var buf strings.Builder
	logger := zerolog.New(&buf)

	_ = Migrate("postgres://alice:super-secret-pw@db.internal:5432/rvpay?sslmode=require", "/path/that/does/not/exist", logger)

	output := buf.String()
	if strings.Contains(output, "super-secret-pw") {
		t.Errorf("migration logs leaked the database password: %q", output)
	}
	if strings.Contains(output, "postgres://alice:") {
		t.Errorf("migration logs leaked the database URL: %q", output)
	}
}
