package transcoder

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/pkg/aws"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"
)

func TestS3Downloader_Download(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	l := logger.New(cfg.Log)

	data := utils.RandString()
	filePath := fmt.Sprintf("s3://local-bucket/%s", utils.RandString())
	require.NoError(t, aws.Write(ctx, cfg.AWS, filePath, []byte(data)))

	for _, te := range []struct {
		name           string
		filePath       string
		expectedErrStr string
	}{
		{
			name:     "valid_file",
			filePath: filePath,
		},
		{
			name:           "invalid_path",
			filePath:       "invalid-path",
			expectedErrStr: aws.ErrInvalidS3URL.Error(),
		},
		{
			name:           "not_exist_file",
			filePath:       "s3://local-bucket/not-exist-file",
			expectedErrStr: "NoSuchKey",
		},
		{
			name:           "invalid_bucket",
			filePath:       "s3://invalid-bucket/test-file",
			expectedErrStr: "The specified bucket does not exist",
		},
	} {
		t.Run(te.name, func(t *testing.T) {
			downloader := newS3Downloader(l, cfg.AWS)
			outputFolder, outputFilePath, clean, err := downloader.Download(ctx, te.filePath)
			if te.expectedErrStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), te.expectedErrStr)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, outputFilePath)
			require.True(t, strings.HasPrefix(outputFilePath, outputFolder))
			file, err := os.Open(outputFilePath)
			require.NoError(t, err)
			defer file.Close()
			buf := bytes.NewBuffer(nil)
			_, err = buf.ReadFrom(file)
			require.NoError(t, err)
			require.Equal(t, data, string(buf.Bytes()))
			require.NoError(t, file.Close())
			clean()
			require.NoFileExists(t, outputFilePath)
		})
	}
}
