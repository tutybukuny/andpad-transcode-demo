package watcher

import (
	"context"
	"time"

	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/services"
	"transcode-demo/internal/services/transcoderequestsvc"
	"transcode-demo/pkg/utils"
)

type Watcher struct {
	l           *z.Logger
	cfg         *config.Config
	transReqSvc services.ITranscodeRequestService
}

func NewWatcher(cfg *config.Config, l *z.Logger, db *gorm.DB) *Watcher {
	return &Watcher{
		cfg:         cfg,
		l:           l,
		transReqSvc: transcoderequestsvc.NewService(cfg, db, l),
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	w.l.Info("starting watcher")
	ticker := time.NewTicker(w.cfg.WatcherSleepInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			w.l.Info("watcher stopped")
			return nil
		case <-ticker.C:
			err := utils.Recover(w.l, func() error {
				err := w.transReqSvc.CheckTranscodeRequest(ctx)
				if err != nil {
					w.l.Error("failed to check transcode request", z.Error(err))
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
			ticker.Reset(w.cfg.WatcherSleepInterval)
		}
	}
}
