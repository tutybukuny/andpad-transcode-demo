package repositories

import (
	"context"
	"time"

	"transcode-demo/internal/models/entity"
	"transcode-demo/pkg/db"
)

type ITranscodeRequestRepo interface {
	db.IInsert[entity.TranscodeRequest, int64]
	db.IUpdate[entity.TranscodeRequest, int64]
	db.IDelete[entity.TranscodeRequest, int64]
	db.IFindByID[entity.TranscodeRequest, int64]

	GetStalledProcessingRequests(ctx context.Context, hangDuration time.Duration, limit int) ([]entity.TranscodeRequest, error)
	UpdateLastProcessingAt(ctx context.Context, reqID int64, lastProcessingAt time.Time) error
}
