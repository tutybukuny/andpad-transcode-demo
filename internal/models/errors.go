package models

import (
	"errors"

	"transcode-demo/pkg/cerrors"
)

var (
	ErrModelNotFound                    = errors.New("model not found")
	ErrFileNotFound                     = errors.New("file not found")
	ErrTranscodeRequestNotFound         = cerrors.Error(404001, "transcode request not found")
	ErrUnexpectedTranscodeRequestStatus = cerrors.Error(400002, "unexpected transcode request status")
	ErrTranscodeRequestCancelled        = cerrors.Error(400003, "transcode request cancelled")
)
