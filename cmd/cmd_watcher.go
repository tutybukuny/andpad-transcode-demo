package main

import (
	"context"

	"github.com/spf13/cobra"
	z "go.uber.org/zap"

	"transcode-demo/config"
	"transcode-demo/internal/presentation/watcher"
	"transcode-demo/pkg/utils"
)

func newWatcherCmd(cfg *config.Config, l *z.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watcher",
		Short: "Watching the data",
		Long:  "Watching the data and control the reprocessing",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			gormDB, closeConn, err := getDB(cfg, l)
			if err != nil {
				l.Error("Failed to connect to database", z.Error(err))
				return err
			}
			defer closeConn()

			w := watcher.NewWatcher(cfg, l, gormDB)
			utils.WaitShutDown(ctx, l, func() error {
				cancel()
				return nil
			})
			err = w.Run(ctx)
			if err != nil {
				l.Error("Failed to start watcher", z.Error(err))
				return err
			}

			return nil
		},
	}

	return cmd
}
