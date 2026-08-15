package repo

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestMigrateInvalidPath(t *testing.T) {
	t.Parallel()

	err := Migrate("postgres://user:pass@localhost:5432/db?sslmode=disable", "/nonexistent/migrations/path", zerolog.Nop())
	if err == nil {
		t.Fatal("expected error for invalid migration path, got nil")
	}
}

func TestMigrateDownInvalidPath(t *testing.T) {
	t.Parallel()

	err := MigrateDown("postgres://user:pass@localhost:5432/db?sslmode=disable", "/nonexistent/migrations/path", zerolog.Nop())
	if err == nil {
		t.Fatal("expected error for invalid migration path, got nil")
	}
}

func TestIntegrationsRepoImplementsInterface(t *testing.T) {
	t.Parallel()

	// Compile-time check that *Impl satisfies IntegrationsRepo.
	var _ IntegrationsRepo = (*Impl)(nil)
}
