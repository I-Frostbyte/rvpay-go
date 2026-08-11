// Package database provides PostgreSQL connection helpers shared by RVPay
// services: a DSN builder, a pool constructor with eager ping, and a
// golang-migrate runner.
//
// Responsibility: construct a pgxpool from configuration values, verify the
// connection, and apply file-based migrations using golang-migrate.
//
// Consumers: service bootstrap code (cmd/grpc-service).
//
// Non-responsibilities: it does not own SQL queries, sqlc code, migrations,
// or repository implementation. Migration files remain owned by each service.
package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // File source for migrations
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// PostgresURL builds a postgres DSN from configuration values. It sets
// sslmode to "disable" when tlsDisabled is true, otherwise "require".
func PostgresURL(dbUser, dbPassword string, dbPort int, dbHost, dbName string, tlsDisabled bool) string {
	queryValues := url.Values{}
	if tlsDisabled {
		queryValues.Add("sslmode", "disable")
	} else {
		queryValues.Add("sslmode", "require")
	}

	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(dbUser, dbPassword),
		Host:     fmt.Sprintf("%s:%d", dbHost, dbPort),
		Path:     dbName,
		RawQuery: queryValues.Encode(),
	}

	return dbURL.String()
}

// Connect creates a pgxpool from dbURL and verifies the connection with an
// eager ping.
func Connect(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Migrate applies all up migrations from migrationPath to the database at
// dbURL. It returns nil when there are no migrations to apply.
func Migrate(dbURL string, migrationPath string, logger zerolog.Logger) error {
	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		return err
	}

	logger.Info().Msgf("absolute migration path: %v", absPath)

	// Create a new migration instance with the absolute path.
	m, err := migrate.New(
		"file://"+absPath,
		dbURL,
	)

	logger.Info().Msgf("migration instance %v", m)
	if err != nil {
		return err
	}
	defer m.Close()

	// Apply migrations
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	logger.Info().Msg("Migrations applied successfully")
	return nil
}