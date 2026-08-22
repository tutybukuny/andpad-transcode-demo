package transcoderequestsvc

import (
	"context"
	"fmt"

	z "go.uber.org/zap"

	"transcode-demo/internal/constant"
)

func (s *Service) CheckTranscodeRequest(ctx context.Context) error {
	backToTodoCount := 0
	moveToFailedCount := 0
	for {
		reqs, err := s.transReqRepo.GetStalledProcessingRequests(ctx, s.cfg.StalledRequestDuration, 1000)
		if err != nil {
			return fmt.Errorf("GetStalledProcessingRequests: %w", err)
		}
		if len(reqs) == 0 {
			s.l.Info(
				"no more stalled processing requests",
				z.Int("back_to_todo_count", backToTodoCount),
				z.Int("move_to_failed_count", moveToFailedCount),
			)
			return nil
		}

		for _, req := range reqs {
			nextStatus := constant.TranscodeRequestStatusTodo
			if req.RetriedTimes > s.cfg.RetriedTimesThreshold {
				// over threshold, mark as failed
				moveToFailedCount++
				nextStatus = constant.TranscodeRequestStatusFailed
			} else {
				// increase retry times
				moveToFailedCount++
				req.RetriedTimes++
			}
			req.Status = nextStatus
			err = s.transReqRepo.Update(ctx, &req)
			if err != nil {
				s.l.Debug(
					"failed to update request",
					z.Int64("id", req.ID),
					z.Any("next_status", req.Status),
					z.Error(err),
				)
				return fmt.Errorf("failed to update request: %w", err)
			}
		}
	}
}
