package transcodesvc

import (
	"context"
	"errors"
	"fmt"
	"time"

	z "go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"transcode-demo/internal/constant"
	"transcode-demo/internal/models"
	"transcode-demo/internal/models/entity"
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

	trCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	eg, gCtx := errgroup.WithContext(trCtx)
	eg.Go(func() error {
		defer cancel()
		wErr := s.watchCancelRequest(gCtx, reqID, cancel)
		if wErr != nil {
			s.l.Error("watchCancelRequest failed", z.Error(wErr))
			return fmt.Errorf("watch cancel request failed: %w", wErr)
		}
		return nil
	})

	eg.Go(func() error {
		defer cancel()
		tErr := s.transcodeRequest(gCtx, req)
		if tErr != nil {
			s.l.Error("transcodeRequest failed", z.Error(tErr))
			return fmt.Errorf("transcode request failed: %w", tErr)
		}
		return nil
	})

	return nil
}

func (s *Service) watchCancelRequest(ctx context.Context, reqID int64, cancel context.CancelFunc) error {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			req, err := s.transReqRepo.FindByID(ctx, reqID)
			if err != nil {
				return err
			}
			if req.Status == constant.TranscodeRequestStatusCancelled {
				return models.ErrTranscodeRequestCancelled
			}
			ticker.Reset(5 * time.Second)
		}
	}
}

func (s *Service) transcodeRequest(ctx context.Context, req *entity.TranscodeRequest) error {
	var outputFolder, masterFile string
	outputFolder, masterFile, err := s.transcoder.Transcode(ctx, req)
	switch {
	case errors.As(err, &models.ErrTranscodeRequestCancelled):
		s.l.Info(fmt.Sprintf("transcode request %d cancelled", req.ID))
		return nil
	case err != nil:
		s.l.Error("transcode failed", z.Error(err))
		req.Status = constant.TranscodeRequestStatusFailed
		err = s.transReqRepo.Update(ctx, req)
		if err != nil {
			return cerrors.ErrInternal(err)
		}
		return err
	}

	req.Status = constant.TranscodeRequestStatusCompleted
	req.MasterFileURL = masterFile
	req.OutputURL = outputFolder
	err = s.transReqRepo.Update(ctx, req)
	if err != nil {
		return cerrors.ErrInternal(err)
	}
	return nil
}
