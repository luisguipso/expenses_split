package migrate

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql
var migrationsFS embed.FS

func Run(databaseURL string) error {
	slog.Info("migrate: starting database migrations")

	source, err := iofs.New(migrationsFS, "sql")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, verr := m.Version()
	if verr != nil && !errors.Is(verr, migrate.ErrNilVersion) {
		slog.Warn("migrate: could not read current version", "error", verr)
	} else if errors.Is(verr, migrate.ErrNilVersion) {
		slog.Info("migrate: no migrations applied yet")
	} else {
		slog.Info("migrate: current version", "version", version, "dirty", dirty)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("migrate: already up to date")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	newVersion, _, _ := m.Version()
	slog.Info("migrate: applied successfully", "version", newVersion)
	return nil
}
