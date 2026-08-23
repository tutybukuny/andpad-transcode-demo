package transcoder

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/pkg/aws"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"
)

func TestUploader(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	l := logger.New(cfg.Log)

	t.Run("upload_to_s3", func(t *testing.T) {
		data := utils.RandString()
		filePath := fmt.Sprintf("s3://local-bucket/%s", utils.RandString())
		destPath := fmt.Sprintf("s3://local-bucket/%s", utils.RandString())
		require.NoError(t, aws.Write(ctx, cfg.AWS, filePath, []byte(data)))
		reader, err := aws.OpenReader(ctx, cfg.AWS, filePath)
		require.NoError(t, err)

		uploader, err := NewUploader(l, cfg.AWS, reader, destPath)
		require.NoError(t, err)
		err = uploader.Upload(ctx)
		require.NoError(t, err)
		rData, err := aws.Read(ctx, cfg.AWS, destPath)
		require.NoError(t, err)
		require.Equal(t, data, string(rData))
	})

	t.Run("not_handle_dest_case", func(t *testing.T) {
		_, err := NewUploader(l, cfg.AWS, nil, "/User/test")
		require.ErrorContains(t, err, "unsupported dest url")
	})

	t.Run("not_exist_bucket", func(t *testing.T) {
		uploader, err := NewUploader(l, cfg.AWS, bytes.NewReader(nil), "s3://invalid-bucket/test")
		require.NoError(t, err)
		err = uploader.Upload(ctx)
		require.ErrorContains(t, err, "The specified bucket does not exist")
	})
}
