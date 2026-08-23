package transcoder

import (
	"context"

	"transcode-demo/internal/models/entity"
)

type ITranscoder interface {
	Transcode(ctx context.Context, req *entity.TranscodeRequest) (string, string, error)
}
