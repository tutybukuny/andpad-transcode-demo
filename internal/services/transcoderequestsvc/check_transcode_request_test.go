package transcoderequestsvc

import (
	"context"
	"testing"
	"time"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/models/entity"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"

	"github.com/stretchr/testify/require"
)

func TestService_CheckTranscodeRequest(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	db := config.MustNewGorm(t)
	l := logger.New(cfg.Log)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	mockRepo := &mockTranscodeRequestRepo{
		TranscodeRequestRepo: transReqRepo,
	}

	t.Run("empty", func(t *testing.T) {
		svc := NewService(cfg, db, l)
		err := svc.CheckTranscodeRequest(ctx)
		require.NoError(t, err)
	})

	t.Run("internal_server_error", func(t *testing.T) {
		svc := NewService(cfg, db, l)
		svc.transReqRepo = mockRepo
		require.ErrorContains(t, svc.CheckTranscodeRequest(ctx), "GetStalledProcessingRequests")
	})

	t.Run("has_stalled_requests", func(t *testing.T) {
		iCfg := config.MustNewDevConfig(t)
		iCfg.StalledRequestBatch = 1

		var reqs []*entity.TranscodeRequest
		for idx := range 2 {
			req := &entity.TranscodeRequest{
				VideoURL:         utils.RandString(),
				Status:           constant.TranscodeRequestStatusProcessing,
				LastProcessingAt: new(time.Now().Add(-10 * time.Minute)),
			}
			if idx == 1 {
				req.RetriedTimes = 3
			}
			require.NoError(t, transReqRepo.Insert(ctx, req))
			reqs = append(reqs, req)
			t.Cleanup(func() {
				err := transReqRepo.Delete(ctx, req.ID)
				require.NoError(t, err)
			})
		}

		svc := NewService(cfg, db, l)
		err := svc.CheckTranscodeRequest(ctx)
		require.NoError(t, err)

		var r *entity.TranscodeRequest
		for idx, req := range reqs {
			r, err = transReqRepo.FindByID(ctx, req.ID)
			require.NoError(t, err)
			if idx == 0 {
				require.Equal(t, constant.TranscodeRequestStatusTodo, r.Status)
				require.Equal(t, 1, r.RetriedTimes)
			} else {
				require.Equal(t, constant.TranscodeRequestStatusFailed, r.Status)
				require.Equal(t, req.RetriedTimes, r.RetriedTimes)
			}
		}
	})
}
