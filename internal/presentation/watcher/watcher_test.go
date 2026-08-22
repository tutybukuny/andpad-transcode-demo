package watcher

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/models/entity"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"
)

func TestWatcher_Run(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	cfg.WatcherSleepInterval = 50 * time.Millisecond
	db := config.MustNewGorm(t)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	l := logger.New(cfg.Log)

	for _, te := range []struct {
		name                 string
		fixtureReq           *entity.TranscodeRequest
		expectedStatus       constant.TranscodeRequestStatus
		expectedRetriedTimes int
		expectedFailedReason string
	}{
		{
			name: "no_update",
			fixtureReq: &entity.TranscodeRequest{
				VideoURL:           utils.RandString(),
				Status:             constant.TranscodeRequestStatusProcessing,
				StartedTranscodeAt: new(time.Now().Add(-10 * time.Minute)),
				LastProcessingAt:   new(time.Now().Add(-5 * time.Second)),
			},
			expectedStatus:       constant.TranscodeRequestStatusProcessing,
			expectedRetriedTimes: 0,
		},
		{
			name: "move_back_to_todo",
			fixtureReq: &entity.TranscodeRequest{
				VideoURL:           utils.RandString(),
				Status:             constant.TranscodeRequestStatusProcessing,
				StartedTranscodeAt: new(time.Now().Add(-10 * time.Minute)),
				LastProcessingAt:   new(time.Now().Add(-15 * time.Second)),
			},
			expectedStatus:       constant.TranscodeRequestStatusTodo,
			expectedRetriedTimes: 1,
		},
		{
			name: "move_to_failed_after_retry_limit",
			fixtureReq: &entity.TranscodeRequest{
				VideoURL:           utils.RandString(),
				Status:             constant.TranscodeRequestStatusProcessing,
				StartedTranscodeAt: new(time.Now().Add(-10 * time.Minute)),
				LastProcessingAt:   new(time.Now().Add(-15 * time.Second)),
				RetriedTimes:       3,
			},
			expectedStatus:       constant.TranscodeRequestStatusFailed,
			expectedRetriedTimes: 3,
			expectedFailedReason: "exceed retry times threshold",
		},
	} {
		t.Run(te.name, func(t *testing.T) {
			err := transReqRepo.Insert(ctx, te.fixtureReq)
			require.NoError(t, err)
			t.Cleanup(func() {
				err = transReqRepo.Delete(ctx, te.fixtureReq.ID)
				require.NoError(t, err)
			})

			w := NewWatcher(cfg, l, db)
			iCtx, cancel := context.WithCancel(ctx)
			t.Cleanup(cancel)
			go w.Run(iCtx)
			time.Sleep(100 * time.Millisecond)
			r, err := transReqRepo.FindByID(ctx, te.fixtureReq.ID)
			require.NoError(t, err)
			require.Equal(t, te.expectedStatus, r.Status)
			require.Equal(t, te.expectedRetriedTimes, r.RetriedTimes)
			require.Equal(t, te.expectedFailedReason, r.FailedReason)
		})
	}
}
