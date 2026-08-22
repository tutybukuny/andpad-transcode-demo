package transcoder

import (
	"context"
	"testing"
	"transcode-demo/config"

	"github.com/stretchr/testify/require"
)

func TestS3Downloader_Download(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	client, err := cfg.AWS.GetClient(ctx)
	require.NoError(t, err)
	client.Put
}
