package migrator

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
)

const sourceURL = "file://internal/storage/pg/migrations"

func Run(dsn string) error {
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Log.Info("No migrations to apply")
			return nil
		}
		return err
	}
	logger.Log.Info("All migrations applied successfully")
	return nil
}
