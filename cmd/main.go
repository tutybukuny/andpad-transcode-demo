package main

import (
	z "go.uber.org/zap"

	"transcode-demo/config"
	"transcode-demo/pkg/logger"
)

func main() {
	cfg := config.New()
	l := logger.New(cfg.Log)

	if err := newRootCmd(cfg, l).Execute(); err != nil {
		l.Error("Failed to execute root command", z.Error(err))
	}
}
