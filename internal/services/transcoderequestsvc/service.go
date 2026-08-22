package transcoderequestsvc

import (
	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/repositories/interfaces"
	"transcode-demo/internal/repositories/psqlrepo"
)

type Service struct {
	l            *z.Logger
	cfg          *config.Config
	db           *gorm.DB
	transReqRepo interfaces.ITranscodeRequestRepo
}

func NewService(cfg *config.Config, db *gorm.DB, l *z.Logger) *Service {
	return &Service{
		l:            l,
		cfg:          cfg,
		db:           db,
		transReqRepo: psqlrepo.NewTranscodeRequestRepo(db),
	}
}
