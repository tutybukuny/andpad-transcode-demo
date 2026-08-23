package transcoder

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/internal/models/entity"
	"transcode-demo/pkg/aws"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"
)

func TestTranscoder_execFfmpeg(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	l := logger.New(cfg.Log)
	tr := NewTranscoder(l, cfg)
	data, err := aws.Read(ctx, cfg.AWS, utils.TestMp4File)
	require.NoError(t, err)
	inputFolder := filepath.Join(t.TempDir(), "input")
	inputFile := path.Join(inputFolder, "input.mp4")
	require.NoError(t, os.MkdirAll(inputFolder, 0755))
	require.NoError(t, os.WriteFile(inputFile, data, 0644))

	outputFolder := fmt.Sprintf("%s/output", t.TempDir())
	require.NoError(t, tr.execFfmpeg(ctx, outputFolder, inputFile))

	for _, file := range []string{"output_1080p.m3u8", "output_720p.m3u8", "output_480p.m3u8", "master.m3u8"} {
		_, err = os.Stat(path.Join(outputFolder, file))
		require.NoError(t, err)
	}
}

func TestTranscoder_uploadToStorage(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	l := logger.New(cfg.Log)
	tr := NewTranscoder(l, cfg)
	data, err := aws.Read(ctx, cfg.AWS, utils.TestMp4File)

	require.NoError(t, err)
	inputFolder := filepath.Join(t.TempDir(), "input")
	inputFile := path.Join(inputFolder, "input.mp4")
	require.NoError(t, os.MkdirAll(inputFolder, 0755))
	require.NoError(t, os.WriteFile(inputFile, data, 0644))

	outputFolder := fmt.Sprintf("%s/output", t.TempDir())
	require.NoError(t, tr.execFfmpeg(ctx, outputFolder, inputFile))
	var numFiles int
	require.NoError(t, filepath.Walk(outputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			numFiles++
		}

		return nil
	}))

	destKey := fmt.Sprintf("test-transcode/%s", utils.RandString())
	destFolder := fmt.Sprintf("%s/%s", utils.TestBucketURL, destKey)
	destStreamingFolder := fmt.Sprintf("http://localhost:4566/%s/%s", utils.TestBucket, destKey)
	err = tr.uploadToStorage(ctx, outputFolder, destFolder, destStreamingFolder)
	require.NoError(t, err)

	s3Client, err := cfg.AWS.GetClient(ctx)
	require.NoError(t, err)
	s3Objects, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: new(utils.TestBucket),
		Prefix: new(destKey + "/"),
	})
	require.NoError(t, err)
	require.Equal(t, numFiles, len(s3Objects.Contents))
	validateTranscodeOutput(t, ctx, cfg, destStreamingFolder, s3Objects)
}

func TestTranscoder_Transcode(t *testing.T) {
	ctx := context.Background()
	cfg := config.MustNewDevConfig(t)
	l := logger.New(cfg.Log)
	tr := NewTranscoder(l, cfg)
	outputFolder, masterFile, err := tr.Transcode(ctx, &entity.TranscodeRequest{
		ID:       rand.Int64N(10000000000),
		VideoURL: utils.TestMp4File,
	})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(masterFile, outputFolder))
	require.True(t, strings.HasPrefix(outputFolder, cfg.TranscodeStoragePrefix))

	bucket, key, err := aws.ParseS3URL(outputFolder)
	require.NoError(t, err)

	s3Client, err := cfg.AWS.GetClient(ctx)
	require.NoError(t, err)
	s3Objects, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: new(key + "/"),
	})
	require.NoError(t, err)
	validateTranscodeOutput(t, ctx, cfg, strings.ReplaceAll(outputFolder, cfg.TranscodeStoragePrefix, cfg.TranscodeStreamingPrefix), s3Objects)
}

func validateTranscodeOutput(t *testing.T, ctx context.Context, cfg *config.Config, destStreamingFolder string, s3Objects *s3.ListObjectsV2Output) {
	expectedM3u8Files := []string{"master.m3u8", "output_1080p.m3u8", "output_720p.m3u8", "output_480p.m3u8"}
	var m3u8Files []string
	for _, obj := range s3Objects.Contents {
		if strings.HasSuffix(*obj.Key, "m3u8") {
			data, err := aws.Read(ctx, cfg.AWS, fmt.Sprintf("%s/%s", utils.TestBucketURL, *obj.Key))
			require.NoError(t, err)
			for line := range strings.SplitSeq(string(data), "\n") {
				if !strings.HasSuffix(line, ".m3u8") && !strings.HasSuffix(line, ".ts") {
					continue
				}
				require.Contains(t, line, destStreamingFolder)
				s3Path := strings.ReplaceAll(line, "http://localhost:4566/", "s3://")
				_, err = aws.Read(ctx, cfg.AWS, s3Path)
				require.NoError(t, err)
			}
			nameElements := strings.Split(*obj.Key, "/")
			m3u8Files = append(m3u8Files, nameElements[len(nameElements)-1])
		}
	}
	require.ElementsMatch(t, expectedM3u8Files, m3u8Files)
}
