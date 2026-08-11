package repo

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestWrapNotFound(t *testing.T) {
	t.Parallel()

	// A single shared instance for the "preserved" case so identity comparison
	// is meaningful.
	otherErr := errors.New("other")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "nil", want: nil},
		{name: "pgx no rows", err: pgx.ErrNoRows, want: ErrNotFound},
		{name: "other error preserved", err: otherErr, want: otherErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := wrapNotFound(tt.err)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("wrapNotFound(nil) = %v, want nil", got)
				}
				return
			}
			// For a non-mapped error, wrapNotFound must return the exact same
			// error instance, not a copy.
			if got != tt.want {
				t.Fatalf("wrapNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestWrapErrorConstraintMappings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want error
	}{
		{name: "unique violation", code: "23505", want: ErrDuplicate},
		{name: "foreign key violation", code: "23503", want: ErrConstraint},
		{name: "check violation", code: "23514", want: ErrConstraint},
		{name: "unmapped code preserved", code: "22003"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pgErr := &pgconn.PgError{Code: tt.code}
			got := wrapError(pgErr)
			if tt.want == nil {
				if got == nil {
					t.Fatalf("wrapError(%s) = nil, want the raw error preserved", tt.code)
				}
				var unwrapped *pgconn.PgError
				if !errors.As(got, &unwrapped) {
					t.Fatalf("wrapError(%s) = %v, want the original PgError", tt.code, got)
				}
				return
			}
			if !errors.Is(got, tt.want) {
				t.Fatalf("wrapError(%s) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestWrapErrorNilAndNonPg(t *testing.T) {
	t.Parallel()

	if got := wrapError(nil); got != nil {
		t.Fatalf("wrapError(nil) = %v, want nil", got)
	}

	other := errors.New("generic")
	if got := wrapError(other); !errors.Is(got, other) {
		t.Fatalf("wrapError(generic) = %v, want the generic error preserved", got)
	}
}
