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
		return models.ErrTranscodeRequestCancelled
	case constant.TranscodeRequestStatusProcessing:
		// only allowed status
	default:
		return fmt.Errorf("transcode request status %s: %w", req.Status, models.ErrUnexpectedTranscodeRequestStatus)
	}

	req.StartedTranscodeAt = new(time.Now())
	req.LastProcessingAt = new(time.Now())
	err = s.transReqRepo.Update(ctx, req)
	if err != nil {
		return cerrors.ErrInternal(err)
	}

	tCtx, cancel := context.WithTimeout(ctx, s.cfg.TranscodeTimeLimit)
	defer cancel()
	eg, gCtx := errgroup.WithContext(tCtx)
	eg.Go(func() error {
		wErr := s.watchCancelRequest(gCtx, reqID)
		if wErr != nil {
			s.l.Error("watchCancelRequest failed", z.Error(wErr))
			return fmt.Errorf("watch cancel request failed: %w", wErr)
		}
		return nil
	})

	eg.Go(func() error {
		uErr := s.updateLastProcessingAt(gCtx, reqID)
		if uErr != nil {
			s.l.Error("updateLastProcessingAt failed", z.Error(uErr))
			return fmt.Errorf("update last processing at failed: %w", uErr)
		}
		return nil
	})

	var outputFolder, masterFile string
	eg.Go(func() error {
		defer cancel()
		var tErr error
		outputFolder, masterFile, tErr = s.transcoder.Transcode(gCtx, req)
		if tErr != nil {
			s.l.Error("transcodeRequest failed", z.Error(tErr))
			return fmt.Errorf("transcode request failed: %w", tErr)
		}
		return nil
	})
	err = eg.Wait()
	if gCtx.Err() != nil && !errors.Is(gCtx.Err(), context.Canceled) {
		err = errors.Join(err, gCtx.Err())
	}

	return s.handleResult(ctx, reqID, outputFolder, masterFile, err)
}

func (s *Service) watchCancelRequest(ctx context.Context, reqID int64) error {
	ticker := time.NewTicker(s.cfg.WatchTransReqInterval)
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
			ticker.Reset(s.cfg.WatchTransReqInterval)
		}
	}
}

func (s *Service) updateLastProcessingAt(ctx context.Context, reqID int64) error {
	timer := time.NewTimer(s.cfg.UpdateLastProcessingAtInterval)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
			uErr := s.transReqRepo.UpdateLastProcessingAt(ctx, reqID, time.Now())
			if uErr != nil {
				s.l.Error("updateLastProcessingAt failed", z.Error(uErr))
				return fmt.Errorf("update last processing at failed: %w", uErr)
			}
			timer.Reset(s.cfg.UpdateLastProcessingAtInterval)
		}
	}
}

func (s *Service) handleResult(ctx context.Context, reqID int64, outputFolder, masterFile string, err error) error {
	req, fErr := s.transReqRepo.FindByID(ctx, reqID)
	if fErr != nil {
		return fmt.Errorf("get req before update result: %w", err)
	}
	var cErr *cerrors.CError
	switch {
	case errors.As(err, &cErr):
		if cErr.Code == models.ErrTranscodeRequestCancelled.Code {
			s.l.Info(fmt.Sprintf("transcode request %d cancelled", req.ID))
			req.StoppedAt = new(time.Now())
			break
		}
		fallthrough
	case err != nil:
		s.l.Error("transcode failed", z.Error(err))
		req.Status = constant.TranscodeRequestStatusFailed
		req.FailedReason = err.Error()
	default:
		req.Status = constant.TranscodeRequestStatusCompleted
		req.StoppedAt = new(time.Now())
		req.MasterFileURL = masterFile
		req.OutputURL = outputFolder
	}

	err = s.transReqRepo.Update(ctx, req)
	if err != nil {
		return cerrors.ErrInternal(err)
	}
	return nil
}
