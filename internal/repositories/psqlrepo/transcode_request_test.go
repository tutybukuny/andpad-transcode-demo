package psqlrepo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/models/entity"
	"transcode-demo/pkg/utils"
)

func TestNewTranscodeRequestRepo(t *testing.T) {
	ctx := context.Background()
	db := config.MustNewGorm(t)
	repo := NewTranscodeRequestRepo(db)

	now := time.Now()

	// create transcode request
	req := &entity.TranscodeRequest{
		VideoURL: utils.RandString(),
		Status:   "",
	}
	err := repo.Insert(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, req.ID)
	require.GreaterOrEqual(t, req.CreatedAt, now)
	require.GreaterOrEqual(t, req.UpdatedAt, now)
	t.Logf("created transcode request %d", req.ID)

	// get transcode request
	dbReq, err := repo.FindByID(ctx, req.ID)
	require.NoError(t, err)
	require.Equal(t, req.ID, dbReq.ID)
	require.Equal(t, req.VideoURL, dbReq.VideoURL)
	require.Equal(t, req.OutputURL, dbReq.OutputURL)
	require.Equal(t, constant.TranscodeRequestStatusTodo, dbReq.Status)
	require.Nil(t, dbReq.StartedTranscodeAt)
	require.Nil(t, dbReq.FinishedTranscodeAt)
	require.Equal(t, req.CreatedAt.Unix(), dbReq.CreatedAt.Unix())
	require.Equal(t, req.UpdatedAt.Unix(), dbReq.UpdatedAt.Unix())

	// update transcode request
	outputURL := utils.RandString()
	startedTranscodeAt := now.Add(-10 * time.Minute)
	finishedTranscodeAt := now.Add(10 * time.Minute)
	dbReq.Status = constant.TranscodeRequestStatusCompleted
	dbReq.StartedTranscodeAt = &startedTranscodeAt
	dbReq.FinishedTranscodeAt = &finishedTranscodeAt
	dbReq.OutputURL = outputURL
	err = repo.Update(ctx, dbReq)
	require.NoError(t, err)
	dbReq, err = repo.FindByID(ctx, dbReq.ID)
	require.NoError(t, err)
	require.Equal(t, constant.TranscodeRequestStatusCompleted, dbReq.Status)
	require.Equal(t, outputURL, dbReq.OutputURL)
	require.Equal(t, startedTranscodeAt.Unix(), dbReq.StartedTranscodeAt.Unix())
	require.Equal(t, finishedTranscodeAt.Unix(), dbReq.FinishedTranscodeAt.Unix())

	// reinsert, make sure no duplicate id
	req.ID = 0
	err = repo.Insert(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, req.ID)
	require.Greater(t, req.ID, dbReq.ID)

	// delete transcode request
	for _, id := range []int64{req.ID, dbReq.ID} {
		err = repo.Delete(ctx, id)
		require.NoError(t, err)
	}
}
