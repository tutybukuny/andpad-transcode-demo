package watcher

import (
	"context"
	"time"

	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/services"
	"transcode-demo/internal/services/transcoderequestsvc"
)

type Watcher struct {
	l           *z.Logger
	cfg         *config.Config
	db          *gorm.DB
	transReqSvc services.ITranscodeRequestService
}

func NewWatcher(cfg *config.Config, l *z.Logger, db *gorm.DB) *Watcher {
	return &Watcher{
		cfg:         cfg,
		l:           l,
		db:          db,
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
			err := w.transReqSvc.CheckTranscodeRequest(ctx)
			if err != nil {
				w.l.Error("failed to check transcode request", z.Error(err))
			}
			ticker.Reset(w.cfg.WatcherSleepInterval)
		}
	}
}
