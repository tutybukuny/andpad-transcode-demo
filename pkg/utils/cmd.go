package utils

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	z "go.uber.org/zap"
)

func WaitShutDown(ctx context.Context, l *z.Logger, shutdownFunc func() error) {
	osCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer stop()
		select {
		case <-osCtx.Done():
			l.Info("Shutting down")
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		done := make(chan error)
		go func() {
			done <- shutdownFunc()
		}()

		select {
		case sErr := <-done:
			if sErr != nil {
				l.Error("Trouble when shutting down", z.Error(sErr))
			}
		case <-shutdownCtx.Done():
			l.Info("Shutdown timed out")
		}
	}()
}
