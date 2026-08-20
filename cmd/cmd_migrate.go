package main

import (
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
	z "go.uber.org/zap"

	"transcode-demo/config"
)

//go:embed _migrations/*.sql
var migrationFS embed.FS

func newMigrateCmd(cfg *config.Config, l *z.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migration",
		Long:  "Migration application",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dbConn, err := cfg.GetDB()
			if err != nil {
				l.Error("Failed to connect to database", z.Error(err))
				return err
			}
			defer dbConn.Close()

			driver, err := postgres.WithInstance(dbConn, &postgres.Config{
				SchemaName: "public",
			})
			if err != nil {
				l.Error("Failed to create postgres driver", z.Error(err))
				return err
			}
			sourceDriver, err := iofs.New(migrationFS, "_migrations")
			if err != nil {
				return err
			}
			m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
			if err != nil {
				l.Error("Failed to create migrate instance", z.Error(err))
				return err
			}
			err = m.Up()
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				l.Error("Failed to migrate database", z.Error(err))
				return err
			}
			l.Info("Database migrated successfully")

			return nil
		},
	}

	return cmd
}
