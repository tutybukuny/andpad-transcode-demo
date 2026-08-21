package transcoderequestsvc

import (
	"context"
	"errors"

	"transcode-demo/internal/constant"
	"transcode-demo/internal/models"
	"transcode-demo/internal/models/entity"
	"transcode-demo/pkg/cerrors"
)

func (s *Service) CreateTranscodeRequest(ctx context.Context, videoURL string) (int64, error) {
	req := &entity.TranscodeRequest{
		VideoURL: videoURL,
		Status:   constant.TranscodeRequestStatusTodo,
	}
	err := s.transReqRepo.Insert(ctx, req)
	if err != nil {
		return 0, cerrors.ErrInternal(err)
	}
	return req.ID, nil
}

func (s *Service) GetTranscodeRequest(ctx context.Context, id int64) (*entity.TranscodeRequest, error) {
	transReq, err := s.transReqRepo.FindByID(ctx, id)
	switch {
	case errors.Is(err, models.ErrModelNotFound):
		return nil, cerrors.Error(404001, "transcode request not found")
	case err != nil:
		return nil, cerrors.ErrInternal(err)
	}
	return transReq, nil
}
