package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ErrInvalidS3URL = fmt.Errorf("invalid s3 url")
)

func ParseS3URL(s3URL string) (bucket, key string, err error) {
	u, err := url.Parse(s3URL)
	if err != nil {
		return "", "", err
	}
	return u.Host, strings.TrimPrefix(u.Path, "/"), nil
}

func Write(ctx context.Context, cfg Config, filePath string, data []byte) error {
	if !strings.HasPrefix(filePath, "s3://") {
		return ErrInvalidS3URL
	}
	s3Client, err := cfg.GetClient(ctx)
	if err != nil {
		return err
	}
	bucket, key, err := ParseS3URL(filePath)
	if err != nil {
		return fmt.Errorf("invalid s3 url: %w", err)
	}
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(data),
	})
	return err
}

func Read(ctx context.Context, cfg Config, filePath string) ([]byte, error) {
	if !strings.HasPrefix(filePath, "s3://") {
		return nil, ErrInvalidS3URL
	}
	s3Client, err := cfg.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	bucket, key, err := ParseS3URL(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid s3 url: %w", err)
	}
	res, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body := bytes.NewBuffer(nil)
	_, err = body.ReadFrom(res.Body)
	return body.Bytes(), err
}

func OpenReader(ctx context.Context, cfg Config, filePath string) (io.ReadCloser, error) {
	if !strings.HasPrefix(filePath, "s3://") {
		return nil, ErrInvalidS3URL
	}
	s3Client, err := cfg.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	bucket, key, err := ParseS3URL(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid s3 url: %w", err)
	}
	res, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
