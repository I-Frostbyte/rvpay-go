package repo

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // File source for migrations
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type IntegrationsRepo interface {
	Begin(ctx context.Context) (sqlc.Querier, pgx.Tx, error)
	Do() sqlc.Querier
}

type Impl struct {
	db *pgxpool.Pool
}

func NewIntegrationsRepo(db *pgxpool.Pool) *Impl {
	return &Impl{db: db}
}

func (i *Impl) Begin(ctx context.Context) (sqlc.Querier, pgx.Tx, error) {
	tx, err := i.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	return sqlc.New(tx), tx, nil
}

func (i *Impl) Do() sqlc.Querier {
	return sqlc.New(i.db)
}

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

// This function rolls back migrations from the db.
func MigrateDown(dbURL string, migrationPath string, logger zerolog.Logger) error {
	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		return err
	}

	// Create a new migration instance with the absolute path
	m, err := migrate.New(
		"file://"+absPath,
		dbURL,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	// Apply migrations
	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	logger.Info().Msg("Migrations applied successfully")
	return nil
}