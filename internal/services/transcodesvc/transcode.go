package transcodesvc

import (
	"context"
	"errors"
	"fmt"

	"transcode-demo/internal/constant"
	"transcode-demo/internal/models"
	"transcode-demo/internal/services/transcodesvc/transcoder"
	"transcode-demo/pkg/cerrors"
)

func (s *Service) Transcode(ctx context.Context, reqID int64) error {
	req, err := s.transReqRepo.FindByID(ctx, reqID)
	switch {
	case errors.Is(err, models.ErrModelNotFound):
		return models.ErrTranscodeRequestNotFound
	case err != nil:
		return cerrors.ErrInternal(err)
	}

	switch req.Status {
	case constant.TranscodeRequestStatusCompleted, constant.TranscodeRequestStatusFailed:
		s.l.Info(fmt.Sprintf("transcode request %d already completed", reqID))
		return nil
	case constant.TranscodeRequestStatusCancelled:
		s.l.Info(fmt.Sprintf("transcode request %d already cancelled", reqID))
		return nil
	case constant.TranscodeRequestStatusTodo:
		// only allowed status
	default:
		return fmt.Errorf("transcode request status %s: %w", req.Status, models.ErrUnexpectedTranscodeRequestStatus)
	}

	// todo: transcode here
	_ = transcoder.NewTranscoder(s.l, s.cfg, req)

	return nil
}
