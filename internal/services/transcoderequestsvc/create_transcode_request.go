package transcoderequestsvc

import (
	"context"
	"transcode-demo/internal/constant"
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
