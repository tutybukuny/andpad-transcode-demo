package psqlrepo

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/models"
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

func TestTranscodeRequestRepo_GetStalledProcessingRequests(t *testing.T) {
	ctx := context.Background()
	db := config.MustNewGorm(t)
	repo := NewTranscodeRequestRepo(db)

	// fixture 5 records
	for i := range 5 {
		req := &entity.TranscodeRequest{
			VideoURL:           utils.RandString(),
			Status:             constant.TranscodeRequestStatusProcessing,
			StartedTranscodeAt: new(time.Now().Add(-20 * time.Minute)),
			LastProcessingAt:   new(time.Now().Add(-time.Duration(i) * time.Minute)),
		}
		err := repo.Insert(ctx, req)
		require.NoError(t, err)
		t.Cleanup(func() {
			err = repo.Delete(ctx, req.ID)
			require.NoError(t, err)
		})
	}

	reqs, err := repo.GetStalledProcessingRequests(ctx, 30*time.Minute, 0)
	require.NoError(t, err)
	require.Empty(t, reqs)

	reqs, err = repo.GetStalledProcessingRequests(ctx, 2*time.Minute, 2)
	require.NoError(t, err)
	require.Len(t, reqs, 2)

	reqs, err = repo.GetStalledProcessingRequests(ctx, 2*time.Minute, 0)
	require.NoError(t, err)
	require.Len(t, reqs, 3)

	for _, req := range reqs {
		require.Equal(t, constant.TranscodeRequestStatusProcessing, req.Status)
		require.True(t, req.LastProcessingAt.Before(time.Now().Add(-2*time.Second)))
	}
}

func TestTranscodeRequestRepo_UpdateLastProcessingAt(t *testing.T) {
	ctx := context.Background()
	db := config.MustNewGorm(t)
	repo := NewTranscodeRequestRepo(db)

	req := &entity.TranscodeRequest{
		VideoURL:           utils.RandString(),
		Status:             constant.TranscodeRequestStatusProcessing,
		StartedTranscodeAt: new(time.Now().Add(-20 * time.Minute)),
		LastProcessingAt:   new(time.Now().Add(-10 * time.Second)),
	}
	err := repo.Insert(ctx, req)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = repo.Delete(ctx, req.ID)
		require.NoError(t, err)
	})
	now := time.Now()
	require.NoError(t, repo.UpdateLastProcessingAt(ctx, req.ID, time.Now()))
	r, err := repo.FindByID(ctx, req.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, r.LastProcessingAt.Unix(), now.Unix())
}

func TestTranscodeRequestRepo_PickRequest(t *testing.T) {
	ctx := context.Background()
	db := config.MustNewGorm(t)
	repo := NewTranscodeRequestRepo(db)

	for range 1000 {
		req := &entity.TranscodeRequest{
			VideoURL: utils.RandString(),
			Status:   constant.TranscodeRequestStatusTodo,
		}
		err := repo.Insert(ctx, req)
		require.NoError(t, err)
		t.Cleanup(func() {
			err = repo.Delete(ctx, req.ID)
		})

		wg := sync.WaitGroup{}
		var count atomic.Int32
		for range 2 {
			wg.Go(func() {
				r, iErr := repo.PickRequest(ctx)
				if iErr != nil {
					require.ErrorIs(t, iErr, models.ErrModelNotFound)
					return
				}
				count.Add(1)
				require.Equal(t, req.ID, r.ID)
				require.Equal(t, constant.TranscodeRequestStatusProcessing, r.Status)
			})
		}
		wg.Wait()
		require.Equal(t, 1, int(count.Load()))
	}
}
