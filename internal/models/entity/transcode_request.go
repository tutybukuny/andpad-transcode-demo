package entity

import (
	"time"

	"transcode-demo/internal/constant"
)

type TranscodeRequest struct {
	ID                  int64                           `json:"id"`
	VideoURL            string                          `json:"video_url"`
	OutputURL           string                          `json:"output_url"`
	Status              constant.TranscodeRequestStatus `json:"status" gorm:"default:todo"`
	FailedReason        string                          `json:"failed_reason"`
	StartedTranscodeAt  *time.Time                      `json:"started_transcode_at"`
	FinishedTranscodeAt *time.Time                      `json:"finished_transcode_at"`
	CreatedAt           time.Time                       `json:"created_at"`
	UpdatedAt           time.Time                       `json:"updated_at"`
}

func (TranscodeRequest) IDField() string {
	return "id"
}

func (TranscodeRequest) TableName() string {
	return "tdapp.transcode_request"
}
