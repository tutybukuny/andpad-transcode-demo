package transcodesvc

import (
	"fmt"

	"github.com/panjf2000/ants/v2"
	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/repositories"
	"transcode-demo/internal/repositories/psqlrepo"
)

type Service struct {
	l            *z.Logger
	cfg          *config.Config
	db           *gorm.DB
	transReqRepo repositories.ITranscodeRequestRepo
	threadPool   *ants.Pool
}

func NewService(cfg *config.Config, l *z.Logger, db *gorm.DB) (*Service, error) {
	threadPool, err := ants.NewPool(10)
	if err != nil {
		return nil, fmt.Errorf("failed to create thread pool: %w", err)
	}
	return &Service{
		cfg:          cfg,
		l:            l,
		db:           db,
		transReqRepo: psqlrepo.NewTranscodeRequestRepo(db),
		threadPool:   threadPool,
	}, nil
}
