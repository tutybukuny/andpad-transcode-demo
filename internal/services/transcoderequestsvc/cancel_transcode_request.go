package transcoderequestsvc

import (
	"context"
	"errors"

	"transcode-demo/internal/constant"
	"transcode-demo/internal/models"
	"transcode-demo/pkg/cerrors"
)

func (s *Service) CancelTranscodeRequest(ctx context.Context, id int64) error {
	transReq, err := s.transReqRepo.FindByID(ctx, id)
	switch {
	case errors.Is(err, models.ErrModelNotFound):
		return models.ErrTranscodeRequestNotFound
	case err != nil:
		return cerrors.ErrInternal(err)
	}

	transReq.Status = constant.TranscodeRequestStatusCancelled
	err = s.transReqRepo.Update(ctx, transReq)
	if err != nil {
		return cerrors.ErrInternal(err)
	}
	return nil
}
