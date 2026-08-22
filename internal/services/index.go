package services

import (
	"context"

	"transcode-demo/internal/models/entity"
)

type ITranscodeRequestService interface {
	CreateTranscodeRequest(ctx context.Context, videoURL string) (int64, error)
	GetTranscodeRequest(ctx context.Context, id int64) (*entity.TranscodeRequest, error)
	CancelTranscodeRequest(ctx context.Context, id int64) error
	CheckTranscodeRequest(ctx context.Context) error
}
