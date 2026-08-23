package transcoder

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	z "go.uber.org/zap"

	"transcode-demo/config"
	"transcode-demo/pkg/aws"
)

var (
	s3Downloader     *S3Downloader
	s3DownloaderOnce sync.Once
)

type IDownloader interface {
	Download(ctx context.Context, url string) (string, func(), error)
}

func NewDownloader(l *z.Logger, cfg *config.Config, videoURL string) (*S3Downloader, error) {
	switch {
	case strings.HasPrefix(videoURL, "s3://"):
		if s3Downloader == nil {
			s3DownloaderOnce.Do(func() {
				s3Downloader = newS3Downloader(l, cfg.AWS)
			})
		}
		return s3Downloader, nil
	}
	return nil, fmt.Errorf("unsupported video url: %s", videoURL)
}

type S3Downloader struct {
	l   *z.Logger
	cfg aws.Config
}

func newS3Downloader(l *z.Logger, cfg aws.Config) *S3Downloader {
	return &S3Downloader{l: l, cfg: cfg}
}

func (s *S3Downloader) Download(ctx context.Context, videoURL string) (string, string, func(), error) {
	folder, err := os.MkdirTemp("", "transcode")
	if err != nil {
		return "", "", nil, err
	}

	data, err := aws.OpenReader(ctx, s.cfg, videoURL)
	if err != nil {
		return "", "", nil, err
	}
	defer data.Close()
	filePath := path.Join(folder, "inputvideo")
	file, err := os.Create(filePath)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	downloadedBytes, err := io.Copy(file, data)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to download file: %w", err)
	}
	s.l.Info("downloaded the file", z.Int64("downloaded_bytes", downloadedBytes))

	return folder, filePath, func() {
		os.RemoveAll(folder)
	}, nil
}
