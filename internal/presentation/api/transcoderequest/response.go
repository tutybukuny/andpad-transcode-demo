package transcoderequest

import "transcode-demo/internal/models/entity"

type CreateTranscodeRequestResp struct {
	RequestID int64 `json:"request_id"`
}

type GetTranscodeRequestResp struct {
	*entity.TranscodeRequest
}
