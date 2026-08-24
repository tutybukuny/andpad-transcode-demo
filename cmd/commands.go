package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/pkg/db"
)

func newRootCmd(cfg *config.Config, l *z.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tdapp",
		Short: "Transcode Demo",
		Long:  "Transcode service demo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newMigrateCmd(cfg, l))
	cmd.AddCommand(newAPICmd(cfg, l))
	cmd.AddCommand(newWatcherCmd(cfg, l))
	cmd.AddCommand(newWorkerCmd(cfg, l))
	return cmd
}

func getDB(cfg *config.Config, l *z.Logger) (*gorm.DB, func(), error) {
	dbConn, err := cfg.GetDB()
	if err != nil {
		l.Error("Failed to connect to database", z.Error(err))
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	gormDB, err := db.NewGormDB(dbConn)
	if err != nil {
		cErr := dbConn.Close()
		if cErr != nil {
			err = errors.Join(err, cErr)
		}
		return nil, nil, fmt.Errorf("failed to create db: %w", err)
	}

	closeFunc := func() {
		iErr := dbConn.Close()
		if iErr != nil {
			l.Error("Failed to close database connection", z.Error(iErr))
		}
	}

	return gormDB, closeFunc, nil
}
