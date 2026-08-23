package transcoder

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"transcode-demo/internal/models/entity"

	z "go.uber.org/zap"

	"transcode-demo/config"
)

type Transcoder struct {
	l   *z.Logger
	cfg *config.Config
}

func NewTranscoder(l *z.Logger, cfg *config.Config) *Transcoder {
	return &Transcoder{
		l:   l,
		cfg: cfg,
	}
}

func (t *Transcoder) Transcode(ctx context.Context, req *entity.TranscodeRequest) (string, string, error) {
	downloader, err := NewDownloader(t.l, t.cfg, req.VideoURL)
	if err != nil {
		return "", "", fmt.Errorf("NewDownloader: %w", err)
	}
	downloadFolder, downloadPath, cleanup, err := downloader.Download(ctx, req.VideoURL)
	if err != nil {
		return "", "", err
	}
	defer cleanup()

	transcodeOutputFolder, err := t.transcode(ctx, req.ID, downloadFolder, downloadPath)
	if err != nil {
		return "", "", err
	}

	return transcodeOutputFolder, transcodeOutputFolder + "/master.m3u8", nil
}

func (t *Transcoder) transcode(ctx context.Context, reqID int64, downloadFolder, downloadPath string) (string, error) {
	localTranscodeOutputFolder := fmt.Sprintf("%s/output", downloadFolder)
	destTranscodeOutputFolder := fmt.Sprintf("%s/%d", t.cfg.TranscodeStoragePrefix, reqID)
	destTranscodeStreamingOutputFolder := fmt.Sprintf("%s/%d", t.cfg.TranscodeStreamingPrefix, reqID)
	t.l.Info(
		"start transcoding",
		z.String("download_path", downloadPath),
		z.String("local_transcode_output_folder", localTranscodeOutputFolder),
		z.String("dest_transcode_output_folder", destTranscodeOutputFolder),
		z.String("dest_transcode_streaming_output_folder", destTranscodeStreamingOutputFolder),
	)

	if err := t.execFfmpeg(ctx, localTranscodeOutputFolder, downloadPath); err != nil {
		return "", fmt.Errorf("execFfmpeg: %w", err)
	}
	t.l.Debug(
		"ffmpeg done",
		z.String("download_path", downloadPath),
		z.String("local_transcode_output_folder", localTranscodeOutputFolder),
		z.String("dest_transcode_output_folder", destTranscodeOutputFolder),
		z.String("dest_transcode_streaming_output_folder", destTranscodeStreamingOutputFolder),
	)

	if err := t.uploadToStorage(ctx, localTranscodeOutputFolder, destTranscodeOutputFolder, destTranscodeStreamingOutputFolder); err != nil {
		return "", fmt.Errorf("uploadToStorage: %w", err)
	}
	t.l.Info(
		"transcoding completed",
		z.String("download_path", downloadPath),
		z.String("local_transcode_output_folder", localTranscodeOutputFolder),
		z.String("dest_transcode_output_folder", destTranscodeOutputFolder),
		z.String("dest_transcode_streaming_output_folder", destTranscodeStreamingOutputFolder),
	)

	return destTranscodeOutputFolder, nil
}

func (t *Transcoder) execFfmpeg(ctx context.Context, outputFolder, inputFile string) error {
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return fmt.Errorf("failed to create output folder: %w", err)
	}

	args := []string{
		"-i", inputFile,
		"-map", "0:v:0", "-map", "0:a:0", "-map", "0:v:0", "-map", "0:a:0", "-map", "0:v:0", "-map", "0:a:0",
		"-c:v", "libx264", "-c:a", "aac",
		"-filter:v:0", "scale=1920:1080", "-b:v:0", "5000k", "-maxrate:v:0", "5350k", "-bufsize:v:0", "7500k", "-b:a:0", "192k",
		"-filter:v:1", "scale=1280:720", "-b:v:1", "2800k", "-maxrate:v:1", "2996k", "-bufsize:v:1", "4200k", "-b:a:1", "128k",
		"-filter:v:2", "scale=854:480", "-b:v:2", "1400k", "-maxrate:v:2", "1498k", "-bufsize:v:2", "2100k", "-b:a:2", "96k",
		"-var_stream_map", "v:0,a:0,name:1080p v:1,a:1,name:720p v:2,a:2,name:480p",
		"-preset", "veryfast", "-hls_time", "6", "-hls_list_size", "0",
		"-hls_flags", "independent_segments",
		"-master_pl_name", "master.m3u8",
		fmt.Sprintf("%s/output_%%v.m3u8", outputFolder),
	}
	output, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput()
	if err != nil {
		t.l.Error("ffmpeg failed", z.String("output", string(output)), z.Error(err))
		return fmt.Errorf("ffmpeg failed: %w", err)
	}
	return nil
}

func (t *Transcoder) uploadToStorage(ctx context.Context, srcFolder, destFolder, destStreaming string) error {
	err := filepath.Walk(srcFolder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		newPath := strings.ReplaceAll(path, srcFolder, destFolder)
		if !strings.HasSuffix(path, ".m3u8") {
			err = t.uploadTSFile(ctx, path, newPath)
			if err != nil {
				return fmt.Errorf("uploadTSFile: %w", err)
			}
			return nil
		}
		err = t.uploadM3u8File(ctx, path, newPath, destStreaming)
		if err != nil {
			return fmt.Errorf("uploadM3u8File: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	return nil
}

func (t *Transcoder) uploadTSFile(ctx context.Context, srcPath, destPath string) error {
	inputFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer inputFile.Close()
	uploader, err := NewUploader(t.l, t.cfg.AWS, inputFile, destPath)
	if err != nil {
		return fmt.Errorf("NewUploader: %w", err)
	}
	return uploader.Upload(ctx)
}

func (t *Transcoder) uploadM3u8File(ctx context.Context, srcPath, destPath, destStreaming string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	patchedData := []byte(strings.ReplaceAll(string(data), "output_", destStreaming+"/output_"))
	uploader, err := NewUploader(t.l, t.cfg.AWS, bytes.NewReader(patchedData), destPath)
	if err != nil {
		return fmt.Errorf("NewUploader: %w", err)
	}

	return uploader.Upload(ctx)
}
