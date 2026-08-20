package transcoderequestsvc

import (
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/repositories/interfaces"
	"transcode-demo/internal/repositories/psqlrepo"
)

type Service struct {
	cfg          *config.Config
	db           *gorm.DB
	transReqRepo interfaces.ITranscodeRequestRepo
}

func NewService(cfg *config.Config, db *gorm.DB) *Service {
	return &Service{
		cfg:          cfg,
		db:           db,
		transReqRepo: psqlrepo.NewTranscodeRequestRepo(db),
	}
}
