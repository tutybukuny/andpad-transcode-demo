package transcodesvc

import (
	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/repositories"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/internal/services/transcodesvc/transcoder"
)

type Service struct {
	l            *z.Logger
	cfg          *config.Config
	db           *gorm.DB
	transReqRepo repositories.ITranscodeRequestRepo
	transcoder   transcoder.ITranscoder
}

func NewService(cfg *config.Config, l *z.Logger, db *gorm.DB) *Service {
	return &Service{
		cfg:          cfg,
		l:            l,
		db:           db,
		transReqRepo: psqlrepo.NewTranscodeRequestRepo(db),
		transcoder:   transcoder.NewTranscoder(l, cfg),
	}
}
