package transcoderequestsvc

import (
	"context"
	"testing"

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

func TestService_CancelTranscodeRequest(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	db := config.MustNewGorm(t)
	l := logger.New(cfg.Log)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	mockRepo := &mockTranscodeRequestRepo{
		TranscodeRequestRepo: transReqRepo,
	}

	fixtureReq := func(t testing.TB, status constant.TranscodeRequestStatus) *entity.TranscodeRequest {
		req := &entity.TranscodeRequest{
			VideoURL: utils.RandString(),
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

	t.Run("success", func(t *testing.T) {
		reqs := []*entity.TranscodeRequest{
			fixtureReq(t, constant.TranscodeRequestStatusTodo),
			fixtureReq(t, constant.TranscodeRequestStatusProcessing),
			fixtureReq(t, constant.TranscodeRequestStatusCancelled),
		}
		svc := NewService(cfg, db, l)
		for _, req := range reqs {
			err := svc.CancelTranscodeRequest(ctx, req.ID)
			require.NoError(t, err)
			r, err := transReqRepo.FindByID(ctx, req.ID)
			require.NoError(t, err)
			require.Equal(t, constant.TranscodeRequestStatusCancelled, r.Status)
		}
	})

	t.Run("already_finished", func(t *testing.T) {
		reqs := []*entity.TranscodeRequest{
			fixtureReq(t, constant.TranscodeRequestStatusCompleted),
			fixtureReq(t, constant.TranscodeRequestStatusFailed),
		}
		svc := NewService(cfg, db, l)
		for _, req := range reqs {
			err := svc.CancelTranscodeRequest(ctx, req.ID)
			cErr := utils.AsCError(t, err)
			require.Equal(t, cerrors.BadRequest, cErr.Code)
			require.Equal(t, "transcode request already completed", cErr.Message)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		svc := NewService(cfg, db, l)
		err := svc.CancelTranscodeRequest(ctx, 0)
		require.ErrorIs(t, err, models.ErrTranscodeRequestNotFound)
	})

	t.Run("internal_server_error", func(t *testing.T) {
		svc := NewService(cfg, db, l)
		svc.transReqRepo = mockRepo
		err := svc.CancelTranscodeRequest(ctx, 0)
		cErr := utils.AsCError(t, err)
		require.Equal(t, cerrors.InternalServerError, cErr.Code)
	})
}
