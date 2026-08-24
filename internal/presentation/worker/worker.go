package worker

import (
	"context"
	"errors"
	"time"

	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/models"
	"transcode-demo/internal/repositories"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/internal/services"
	"transcode-demo/internal/services/transcodesvc"
	"transcode-demo/pkg/utils"
)

type Worker struct {
	l            *z.Logger
	cfg          *config.Config
	transReqRepo repositories.ITranscodeRequestRepo
	transSvc     services.ITranscodeService
}

func NewWorker(cfg *config.Config, l *z.Logger, db *gorm.DB) *Worker {
	return &Worker{
		l:            l,
		cfg:          cfg,
		transReqRepo: psqlrepo.NewTranscodeRequestRepo(db),
		transSvc:     transcodesvc.NewService(cfg, l, db),
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.l.Info("starting worker")
	ticker := time.NewTicker(w.cfg.GetTranscodeRequestInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			w.l.Info("worker stopped")
			return nil
		case <-ticker.C:
			req, err := w.transReqRepo.PickRequest(ctx)
			switch {
			case errors.Is(err, models.ErrModelNotFound):
				w.l.Info("no transcode request found, waiting...")
				ticker.Reset(w.cfg.GetTranscodeRequestInterval)
				continue
			case err != nil:
				w.l.Error("failed to pick request", z.Error(err))
				ticker.Reset(w.cfg.GetTranscodeRequestInterval)
				continue
			}

			err = utils.Recover(w.l, func() error {
				return w.transSvc.Transcode(ctx, req.ID)
			})
			if err != nil {
				w.l.Error("failed to transcode request", z.Int64("req_id", req.ID), z.Error(err))
			}
		}
	}
}
