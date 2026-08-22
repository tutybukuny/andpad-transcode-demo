package main

import (
	"github.com/spf13/cobra"
	z "go.uber.org/zap"

	"transcode-demo/config"
	"transcode-demo/internal/presentation/api"
	"transcode-demo/pkg/utils"
)

func newAPICmd(cfg *config.Config, l *z.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "API",
		Long:  "API application",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			gormDB, closeConn, err := getDB(cfg, l)
			if err != nil {
				l.Error("Failed to connect to database", z.Error(err))
				return err
			}
			defer closeConn()

			s := api.NewServer(cfg, l, gormDB)
			utils.WaitShutDown(cmd.Context(), l, func() error {
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
