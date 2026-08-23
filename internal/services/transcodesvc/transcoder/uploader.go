package transcoder

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	z "go.uber.org/zap"

	"transcode-demo/pkg/aws"
)

var (
	s3Uploader     *S3Uploader
	s3UploaderOnce sync.Once
)

type IUploader interface {
	Upload(ctx context.Context, input io.Reader, destPath string) error
}

type Uploader struct {
	l        *z.Logger
	cfg      aws.Config
	input    io.Reader
	destPath string
	uploader IUploader
}

func NewUploader(l *z.Logger, cfg aws.Config, input io.Reader, destPath string) (*Uploader, error) {
	switch {
	case strings.HasPrefix(destPath, "s3://"):
		if s3Uploader == nil {
			s3UploaderOnce.Do(func() {
				s3Uploader = newS3Uploader(l, cfg)
			})
		}
		return &Uploader{
			l:        l,
			cfg:      cfg,
			input:    input,
			destPath: destPath,
			uploader: s3Uploader,
		}, nil
	}

	return nil, fmt.Errorf("unsupported dest url: %s", destPath)
}

func (u *Uploader) Upload(ctx context.Context) error {
	return u.uploader.Upload(ctx, u.input, u.destPath)
}

type S3Uploader struct {
	l   *z.Logger
	cfg aws.Config
}

func newS3Uploader(l *z.Logger, cfg aws.Config) *S3Uploader {
	return &S3Uploader{l: l, cfg: cfg}
}

func (s *S3Uploader) Upload(ctx context.Context, input io.Reader, destPath string) error {
	data, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	err = aws.Write(ctx, s.cfg, destPath, data)
	return err
}
