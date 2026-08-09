package repo

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrNotFound is returned when a requested record does not exist.
	ErrNotFound = errors.New("record not found")
	// ErrDuplicate is returned when a unique constraint is violated.
	ErrDuplicate = errors.New("record already exists")
	// ErrConstraint is returned for other constraint violations.
	ErrConstraint = errors.New("constraint violation")
)

// wrapError converts a raw persistence error into a repository-level error
// without exposing PostgreSQL-specific details.
func wrapError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrDuplicate
		case "23503": // foreign_key_violation
			return ErrConstraint
		case "23514": // check_violation
			return ErrConstraint
		default:
			return err
		}
	}

	return err
}

// wrapNotFound converts a not-found persistence error into ErrNotFound.
func wrapNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}