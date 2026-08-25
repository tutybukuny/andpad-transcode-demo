package transcodesvc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	z "go.uber.org/zap"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/models"
	"transcode-demo/internal/models/entity"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"
)

type mockTranscoder struct {
	l             *z.Logger
	sleepInterval time.Duration
}

func (m *mockTranscoder) Transcode(ctx context.Context, req *entity.TranscodeRequest) (string, string, error) {
	m.l.Info("mock transcoding start")
	utils.SleepWithContext(ctx, m.sleepInterval)
	m.l.Info("mock transcoding finished")
	return "output_url", "output_key", nil
}

func TestService_Transcode(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	l := logger.New(cfg.Log)
	db := config.MustNewGorm(t)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)

	fixtureReq := func(t testing.TB, status constant.TranscodeRequestStatus) *entity.TranscodeRequest {
		req := &entity.TranscodeRequest{
			VideoURL: utils.TestMp4File,
			Status:   status,
		}
		err := transReqRepo.Insert(ctx, req)
		require.NoError(t, err)
		t.Cleanup(func() {
			err = transReqRepo.Delete(ctx, req.ID)
			require.NoError(t, err)
		})
		return req
	}

	for _, te := range []struct {
		name           string
		status         constant.TranscodeRequestStatus
		expectedStatus constant.TranscodeRequestStatus
		expectedErr    error
		preRunFunc     func(t *testing.T, req *entity.TranscodeRequest)
		assertFunc     func(t *testing.T, req *entity.TranscodeRequest)
	}{
		{
			name:           "request_processing",
			status:         constant.TranscodeRequestStatusProcessing,
			expectedStatus: constant.TranscodeRequestStatusCompleted,
			assertFunc: func(t *testing.T, req *entity.TranscodeRequest) {
				require.NotEmpty(t, req.OutputURL)
			},
		},
		{
			name:           "request_completed",
			status:         constant.TranscodeRequestStatusCompleted,
			expectedStatus: constant.TranscodeRequestStatusCompleted,
		},
		{
			name:           "request_failed",
			status:         constant.TranscodeRequestStatusFailed,
			expectedStatus: constant.TranscodeRequestStatusFailed,
		},
		{
			name:        "request_processing",
			status:      constant.TranscodeRequestStatusTodo,
			expectedErr: models.ErrUnexpectedTranscodeRequestStatus,
		},
		{
			name:   "request_not_found",
			status: constant.TranscodeRequestStatusProcessing,
			preRunFunc: func(t *testing.T, req *entity.TranscodeRequest) {
				err := transReqRepo.Delete(ctx, req.ID)
				require.NoError(t, err)
			},
			expectedErr: models.ErrTranscodeRequestNotFound,
		},
		{
			name:   "request_invalid_url",
			status: constant.TranscodeRequestStatusProcessing,
			preRunFunc: func(t *testing.T, req *entity.TranscodeRequest) {
				req.VideoURL = "invalid_url"
				err := transReqRepo.Update(ctx, req)
				require.NoError(t, err)
			},
			expectedStatus: constant.TranscodeRequestStatusFailed,
		},
	} {
		t.Run(te.name, func(t *testing.T) {
			req := fixtureReq(t, te.status)
			if te.preRunFunc != nil {
				te.preRunFunc(t, req)
			}
			s := NewService(cfg, l, db)
			err := s.Transcode(ctx, req.ID)
			if te.expectedErr != nil {
				require.ErrorIs(t, err, te.expectedErr)
				return
			}
			require.NoError(t, err)
			req, err = transReqRepo.FindByID(ctx, req.ID)
			require.NoError(t, err)
			require.Equal(t, te.expectedStatus, req.Status)
			if te.assertFunc != nil {
				te.assertFunc(t, req)
			}
		})
	}

	t.Run("request_cancelled", func(t *testing.T) {
		req := fixtureReq(t, constant.TranscodeRequestStatusProcessing)
		cCfg := config.MustNewDevConfig(t)
		cCfg.WatchTransReqInterval = 100 * time.Millisecond
		s := NewService(cCfg, l, db)
		s.transcoder = &mockTranscoder{l: l, sleepInterval: 10 * time.Second}
		go func() {
			time.Sleep(200 * time.Millisecond)
			req.Status = constant.TranscodeRequestStatusCancelled
			err := transReqRepo.Update(ctx, req)
			require.NoError(t, err)
		}()
		err := s.Transcode(ctx, req.ID)
		require.NoError(t, err)
		req, err = transReqRepo.FindByID(ctx, req.ID)
		require.NoError(t, err)
		require.Equal(t, constant.TranscodeRequestStatusCancelled, req.Status)
	})

	t.Run("timeout", func(t *testing.T) {
		req := fixtureReq(t, constant.TranscodeRequestStatusProcessing)
		cCfg := config.MustNewDevConfig(t)
		cCfg.TranscodeTimeLimit = 100 * time.Millisecond
		s := NewService(cCfg, l, db)
		err := s.Transcode(ctx, req.ID)
		require.NoError(t, err)
		req, err = transReqRepo.FindByID(ctx, req.ID)
		require.NoError(t, err)
		require.Equal(t, constant.TranscodeRequestStatusFailed, req.Status)
		require.Contains(t, req.FailedReason, context.DeadlineExceeded.Error())
	})

	t.Run("check_update_last_processing_at", func(t *testing.T) {
		req := fixtureReq(t, constant.TranscodeRequestStatusProcessing)
		cCfg := config.MustNewDevConfig(t)
		cCfg.UpdateLastProcessingAtInterval = 100 * time.Millisecond
		s := NewService(cCfg, l, db)
		s.transcoder = &mockTranscoder{l: l, sleepInterval: 10 * time.Second}
		lastVal := time.Now()
		cCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			time.Sleep(200 * time.Millisecond)
			rReq, iErr := transReqRepo.FindByID(cCtx, req.ID)
			require.NoError(t, iErr)
			require.GreaterOrEqual(t, rReq.LastProcessingAt.Unix(), lastVal.Unix())
			lastVal = *rReq.LastProcessingAt
		}()
		err := s.Transcode(ctx, req.ID)
		require.NoError(t, err)
		req, err = transReqRepo.FindByID(ctx, req.ID)
		require.NoError(t, err)
		require.Equal(t, constant.TranscodeRequestStatusCompleted, req.Status)
		require.NotNil(t, req.StartedTranscodeAt)
		require.NotNil(t, req.FinishedTranscodeAt)
	})
}
