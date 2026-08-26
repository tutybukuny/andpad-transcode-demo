package transcoderequestsvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/models"
	"transcode-demo/internal/models/entity"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/pkg/cerrors"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"
)

type mockTranscodeRequestRepo struct {
	*psqlrepo.TranscodeRequestRepo
}

func (r *mockTranscodeRequestRepo) FindByID(_ context.Context, _ int64) (*entity.TranscodeRequest, error) {
	return nil, fmt.Errorf("mock error")
}

func (r *mockTranscodeRequestRepo) Insert(_ context.Context, _ *entity.TranscodeRequest) error {
	return fmt.Errorf("mock error")
}

func (r *mockTranscodeRequestRepo) GetStalledProcessingRequests(_ context.Context, _ time.Duration, _ int) ([]entity.TranscodeRequest, error) {
	return nil, fmt.Errorf("mock error")
}

func TestService_GetTranscodeRequest(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	db := config.MustNewGorm(t)
	l := logger.New(cfg.Log)
	svc := NewService(cfg, db, l)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	mockRepo := &mockTranscodeRequestRepo{
		TranscodeRequestRepo: transReqRepo,
	}

	// fixture a request
	req := &entity.TranscodeRequest{
		VideoURL: utils.RandString(),
		Status:   constant.TranscodeRequestStatusTodo,
	}
	require.NoError(t, transReqRepo.Insert(ctx, req))
	t.Cleanup(func() {
		err := transReqRepo.Delete(ctx, req.ID)
		require.NoError(t, err)
	})

	// get request
	r, err := svc.GetTranscodeRequest(t.Context(), req.ID)
	require.NoError(t, err)
	require.Equal(t, req.ID, r.ID)
	require.Equal(t, req.VideoURL, r.VideoURL)
	require.Equal(t, req.Status, r.Status)

	// delete request and get again
	err = transReqRepo.Delete(ctx, req.ID)
	require.NoError(t, err)
	_, err = svc.GetTranscodeRequest(ctx, req.ID)
	require.ErrorIs(t, err, models.ErrTranscodeRequestNotFound)

	// internal error
	svc.transReqRepo = mockRepo
	_, err = svc.GetTranscodeRequest(ctx, req.ID)
	cErr := utils.AsCError(t, err)
	require.Equal(t, cerrors.InternalServerError, cErr.Code)
}
