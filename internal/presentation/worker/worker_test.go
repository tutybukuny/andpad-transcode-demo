package worker

import (
	"context"
	"testing"
	"time"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/models/entity"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/pkg/utils"

	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/pkg/logger"
)

type mockTranscodeService struct {
}

func (m *mockTranscodeService) Transcode(ctx context.Context, reqID int64) error {
	return nil
}

func TestWorker(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	cfg.GetTranscodeRequestInterval = 100 * time.Millisecond
	db := config.MustNewGorm(t)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	l := logger.New(cfg.Log)

	t.Run("worker_start_stop", func(t *testing.T) {
		w := NewWorker(cfg, l, db)
		tCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
		err := w.Run(tCtx)
		require.NoError(t, err)
	})

	t.Run("recover_from_panic", func(t *testing.T) {
		tCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
		w := NewWorker(cfg, l, db)
		w.transReqRepo = nil
		err := w.Run(tCtx)
		require.NoError(t, err)
	})

	t.Run("pick_and_process_request", func(t *testing.T) {
		var reqs []*entity.TranscodeRequest
		for range 2 {
			req := &entity.TranscodeRequest{
				VideoURL: utils.RandString(),
				Status:   constant.TranscodeRequestStatusTodo,
			}
			err := transReqRepo.Insert(ctx, req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err = transReqRepo.Delete(ctx, req.ID)
			})
			reqs = append(reqs, req)
		}
		w := NewWorker(cfg, l, db)
		w.transSvc = &mockTranscodeService{}
		tCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		err := w.Run(tCtx)
		require.NoError(t, err)
		var r *entity.TranscodeRequest
		for _, req := range reqs {
			r, err = transReqRepo.FindByID(ctx, req.ID)
			require.NoError(t, err)
			require.Equal(t, constant.TranscodeRequestStatusProcessing, r.Status)
		}
	})
}
