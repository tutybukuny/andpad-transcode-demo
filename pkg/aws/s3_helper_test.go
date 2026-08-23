package aws_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/pkg/aws"
	"transcode-demo/pkg/utils"
)

func TestRead(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	data := utils.RandString()
	fileName := utils.RandString()
	filePath := fmt.Sprintf("s3://local-bucket/%s", fileName)

	s3Client, err := cfg.AWS.GetClient(ctx)
	require.NoError(t, err)

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: new("local-bucket"),
		Key:    new(fileName),
		Body:   bytes.NewReader([]byte(data)),
	})
	require.NoError(t, err)

	for _, te := range []struct {
		name           string
		filePath       string
		expectedErrStr string
	}{
		{
			name:     "valid_path",
			filePath: filePath,
		},
		{
			name:           "not_exist_path",
			filePath:       "s3://local-bucket/not-exist-file",
			expectedErrStr: "NoSuchKey",
		},
		{
			name:           "invalid_path",
			filePath:       "invalid-path",
			expectedErrStr: aws.ErrInvalidS3URL.Error(),
		},
		{
			name:           "invalid_bucket",
			filePath:       "s3://invalid-bucket/test-file",
			expectedErrStr: "The specified bucket does not exist",
		},
		{
			name:           "invalid_key",
			filePath:       "s3://local-bucket/",
			expectedErrStr: "input member Key must not be empty",
		},
	} {
		t.Run(te.name, func(t *testing.T) {
			var rData []byte
			rData, err = aws.Read(ctx, cfg.AWS, te.filePath)
			if te.expectedErrStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), te.expectedErrStr)
			} else {
				require.NoError(t, err)
				require.Equal(t, data, string(rData))
			}
		})
	}

	rData, err := aws.Read(ctx, cfg.AWS, filePath)
	require.NoError(t, err)
	require.Equal(t, data, string(rData))
}

func TestWrite(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)

	s3Client, err := cfg.AWS.GetClient(ctx)
	require.NoError(t, err)

	for _, te := range []struct {
		name           string
		filePath       string
		expectedErrStr string
	}{
		{
			name:     "valid_path",
			filePath: "s3://local-bucket/test-file",
		},
		{
			name:     "overwrite_existing_file",
			filePath: "s3://local-bucket/test-file",
		},
		{
			name:           "invalid_path",
			filePath:       "invalid-path",
			expectedErrStr: aws.ErrInvalidS3URL.Error(),
		},
		{
			name:           "invalid_bucket",
			filePath:       "s3://invalid-bucket/test-file",
			expectedErrStr: "The specified bucket does not exist",
		},
		{
			name:           "invalid_key",
			filePath:       "s3://local-bucket/",
			expectedErrStr: "input member Key must not be empty",
		},
	} {
		t.Run(te.name, func(t *testing.T) {
			data := utils.RandString()
			err = aws.Write(ctx, cfg.AWS, te.filePath, []byte(data))
			if te.expectedErrStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), te.expectedErrStr)
			} else {
				require.NoError(t, err)
				var bucket, key string
				bucket, key, err = aws.ParseS3URL(te.filePath)
				require.NoError(t, err)
				var res *s3.GetObjectOutput
				res, err = s3Client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: &bucket,
					Key:    &key,
				})
				require.NoError(t, err)
				defer res.Body.Close()
				buf := new(bytes.Buffer)
				_, err = io.Copy(buf, res.Body)
				require.NoError(t, err)
				require.Equal(t, data, string(buf.Bytes()))
			}
		})
	}
}
