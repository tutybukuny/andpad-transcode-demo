package transcoderequestsvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/pkg/cerrors"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"
)

func TestService_CreateTranscodeRequest(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	db := config.MustNewGorm(t)
	l := logger.New(cfg.Log)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	mockRepo := &mockTranscodeRequestRepo{
		TranscodeRequestRepo: transReqRepo,
	}

	t.Run("success", func(t *testing.T) {
		svc := NewService(cfg, db, l)
		videoURL := utils.RandString()
		id, err := svc.CreateTranscodeRequest(ctx, videoURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err = transReqRepo.Delete(ctx, id)
			require.NoError(t, err)
		})
		require.NotEmpty(t, id)
		req, err := transReqRepo.FindByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, videoURL, req.VideoURL)
		require.Equal(t, constant.TranscodeRequestStatusTodo, req.Status)
	})

	t.Run("internal_server_error", func(t *testing.T) {
		svc := NewService(cfg, db, l)
		svc.transReqRepo = mockRepo
		_, err := svc.CreateTranscodeRequest(ctx, utils.RandString())
		cErr := utils.AsCError(t, err)
		require.Equal(t, cerrors.InternalServerError, cErr.Code)
	})
}
