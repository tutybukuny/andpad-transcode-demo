package models

import (
	"errors"

	"transcode-demo/pkg/cerrors"
)

var (
	ErrModelNotFound            = errors.New("model not found")
	ErrTranscodeRequestNotFound = cerrors.Error(404001, "transcode request not found")
)
