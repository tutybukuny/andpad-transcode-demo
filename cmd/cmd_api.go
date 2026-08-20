package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	z "go.uber.org/zap"

	"transcode-demo/config"
	"transcode-demo/internal/presentation/api"
	"transcode-demo/pkg/db"
	"transcode-demo/pkg/utils"
)

func newAPICmd(cfg *config.Config, l *z.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "API",
		Long:  "API application",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dbConn, err := cfg.GetDB()
			if err != nil {
				l.Error("Failed to connect to database", z.Error(err))
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()
			gormDB, err := db.NewGormDB(dbConn)
			if err != nil {
				return fmt.Errorf("failed to create db: %w", err)
			}

			s := api.NewServer(cfg, l, gormDB)
			utils.WaitShutDown(context.Background(), l, func() error {
				return s.Shutdown()
			})

			if err = s.Listen(cfg.HttpConfig.GetAddr()); err != nil {
				l.Error("Server failed", z.Error(err))
			}

			l.Info("Server shutdown totally")
			return err
		},
	}
	return cmd
}
