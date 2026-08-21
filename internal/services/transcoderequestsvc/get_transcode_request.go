package transcoderequestsvc

import (
	"context"
	"errors"

	"transcode-demo/internal/models"
	"transcode-demo/internal/models/entity"
	"transcode-demo/pkg/cerrors"
)

func (s *Service) GetTranscodeRequest(ctx context.Context, id int64) (*entity.TranscodeRequest, error) {
	transReq, err := s.transReqRepo.FindByID(ctx, id)
	switch {
	case errors.Is(err, models.ErrModelNotFound):
		return nil, models.ErrTranscodeRequestNotFound
	case err != nil:
		return nil, cerrors.ErrInternal(err)
	}
	return transReq, nil
}
