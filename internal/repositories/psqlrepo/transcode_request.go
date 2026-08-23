package psqlrepo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"transcode-demo/internal/constant"
	"transcode-demo/internal/models"
	"transcode-demo/internal/models/entity"
	dbgorm "transcode-demo/pkg/db/gorm"
)

type TranscodeRequestRepo struct {
	*dbgorm.BaseRepo
	*dbgorm.InsertRepo[entity.TranscodeRequest, int64]
	*dbgorm.UpdateRepo[entity.TranscodeRequest, int64]
	*dbgorm.DeleteRepo[entity.TranscodeRequest, int64]
	*dbgorm.FindByIDRepo[entity.TranscodeRequest, int64]
}

func NewTranscodeRequestRepo(db *gorm.DB) *TranscodeRequestRepo {
	base := dbgorm.NewBaseRepo[entity.TranscodeRequest](db)
	return &TranscodeRequestRepo{
		BaseRepo:     base,
		InsertRepo:   dbgorm.NewInsertRepo[entity.TranscodeRequest, int64](base),
		UpdateRepo:   dbgorm.NewUpdateRepo[entity.TranscodeRequest, int64](base),
		DeleteRepo:   dbgorm.NewDeleteRepo[entity.TranscodeRequest, int64](base),
		FindByIDRepo: dbgorm.NewFindByIDRepo[entity.TranscodeRequest, int64](base),
	}
}

func (r *TranscodeRequestRepo) GetStalledProcessingRequests(ctx context.Context, hangDuration time.Duration, limit int) ([]entity.TranscodeRequest, error) {
	query := gorm.G[entity.TranscodeRequest](r.GetDB(ctx)).
		Where(
			"status = ? AND last_processing_at < ?",
			constant.TranscodeRequestStatusProcessing,
			time.Now().Add(-hangDuration),
		)
	if limit > 0 {
		query = query.Limit(limit)
	}
	reqs, err := query.Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stalled processing requests: %w", err)
	}
	return reqs, nil
}

func (r *TranscodeRequestRepo) UpdateLastProcessingAt(ctx context.Context, reqID int64, lastProcessingAt time.Time) error {
	affected, err := gorm.G[entity.TranscodeRequest](r.GetDB(ctx)).Where("id=?", reqID).Update(ctx, "last_processing_at", lastProcessingAt)
	if err != nil {
		return fmt.Errorf("failed to update last processing at: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("failed to update last processing at: %w", models.ErrModelNotFound)
	}
	return nil
}
