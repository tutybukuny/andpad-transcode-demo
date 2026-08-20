package main

import (
	"github.com/spf13/cobra"
	z "go.uber.org/zap"

	"transcode-demo/config"
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
	return cmd
}
