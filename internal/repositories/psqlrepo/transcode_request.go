package psqlrepo

import (
	"transcode-demo/internal/models/entity"
	dbgorm "transcode-demo/pkg/db/gorm"

	"gorm.io/gorm"
)

type TranscodeRequestRepo struct {
	*dbgorm.BaseRepo
	*dbgorm.InsertRepo[entity.TranscodeRequest, int64]
	*dbgorm.UpdateRepo[entity.TranscodeRequest, int64]
	*dbgorm.DeleteRepo[entity.TranscodeRequest, int64]
	*dbgorm.FindByIDRepo[entity.TranscodeRequest, int64]
}

func NewTranscodeRequestRepo(db *gorm.DB) *TranscodeRequestRepo {
	base := dbgorm.NewBaseRepo[entity.TranscodeRequest](db)
	return &TranscodeRequestRepo{
		BaseRepo:     base,
		InsertRepo:   dbgorm.NewInsertRepo[entity.TranscodeRequest, int64](base),
		UpdateRepo:   dbgorm.NewUpdateRepo[entity.TranscodeRequest, int64](base),
		DeleteRepo:   dbgorm.NewDeleteRepo[entity.TranscodeRequest, int64](base),
		FindByIDRepo: dbgorm.NewFindByIDRepo[entity.TranscodeRequest, int64](base),
	}
}
