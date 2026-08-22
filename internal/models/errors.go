package models

import (
	"errors"

	"transcode-demo/pkg/cerrors"
)

var (
	ErrModelNotFound                    = errors.New("model not found")
	ErrFileNotFound                     = errors.New("file not found")
	ErrTranscodeRequestCancelled        = errors.New("transcode request cancelled")
	ErrTranscodeRequestNotFound         = cerrors.Error(404001, "transcode request not found")
	ErrUnexpectedTranscodeRequestStatus = cerrors.Error(400002, "unexpected transcode request status")
)
