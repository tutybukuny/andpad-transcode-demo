package transcoder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	z "go.uber.org/zap"

	"transcode-demo/config"
	"transcode-demo/internal/models"
)

type IDownloader interface {
	Download(ctx context.Context, url string) (string, func(), error)
}

func NewDownloader(ctx context.Context, l *z.Logger, cfg *config.Config, videoURL string) (*S3Downloader, error) {
	switch {
	case strings.HasPrefix(videoURL, "s3://"):
		client, err := cfg.AWS.GetClient(ctx)
		if err != nil {
			return nil, err
		}
		return newS3Downloader(l, client), nil
	}
	return nil, fmt.Errorf("unsupported video url: %s", videoURL)
}

type S3Downloader struct {
	l      *z.Logger
	client *s3.Client
}

func newS3Downloader(l *z.Logger, client *s3.Client) *S3Downloader {
	return &S3Downloader{l: l, client: client}
}

func (s *S3Downloader) Download(ctx context.Context, videoURL string) (string, func(), error) {
	vURL, err := url.Parse(videoURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid video url: %w", err)
	}
	bucket := vURL.Host
	key := vURL.Path

	folder, err := os.MkdirTemp("", "transcode")
	if err != nil {
		return "", nil, err
	}

	res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	var noKey *types.NoSuchKey
	switch {
	case errors.As(err, &noKey):
		return "", nil, fmt.Errorf("video not found: %w", models.ErrFileNotFound)
	case err != nil:
		return "", nil, fmt.Errorf("failed to download video: %w", err)
	}
	defer res.Body.Close()

	filePath := path.Join(folder, "inputvideo")
	file, err := os.Create(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	downloadedBytes, err := io.Copy(file, res.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download file: %w", err)
	}
	s.l.Info("downloaded the file", z.Int64("downloaded_bytes", downloadedBytes))

	return filePath, func() {
		os.RemoveAll(folder)
	}, nil
}
