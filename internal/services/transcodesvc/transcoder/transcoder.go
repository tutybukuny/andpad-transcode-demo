package transcoder

import (
	"context"
	"fmt"
	"transcode-demo/config"
	"transcode-demo/internal/models/entity"

	z "go.uber.org/zap"
)

type ITranscoder interface {
	Start(ctx context.Context) error
	Stop() error
}

type Transcoder struct {
	l   *z.Logger
	cfg *config.Config
	req *entity.TranscodeRequest
}

func NewTranscoder(l *z.Logger, cfg *config.Config, req *entity.TranscodeRequest) *Transcoder {
	return &Transcoder{
		l:   l,
		cfg: cfg,
		req: req,
	}
}

func (t *Transcoder) Transcode(ctx context.Context) (string, error) {
	downloader, err := NewDownloader(ctx, t.l, t.cfg, t.req.VideoURL)
	if err != nil {
		return "", fmt.Errorf("NewDownloader: %w", err)
	}
	downloadPath, cleanup, err := downloader.Download(ctx, t.req.VideoURL)
	if err != nil {
		return "", err
	}
	defer cleanup()

	output, err := t.transcode(ctx, downloadPath)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (t *Transcoder) transcode(ctx context.Context, downloadPath string) (string, error) {
	// todo: transcode
	return "", nil
}
