package dbgorm

import (
	"context"

	"gorm.io/gorm"

	dbconstant "transcode-demo/pkg/db/constant"
	"transcode-demo/pkg/db/entity"
)

type BaseRepo struct {
	db        *gorm.DB
	TableName string
	IDField   string
	ctxDBKey  string
}

func NewBaseRepo[E entity.IEntity](db *gorm.DB, ctxDBKey ...string) *BaseRepo {
	var e E
	dbKey := dbconstant.CtxDBKey
	if len(ctxDBKey) > 0 {
		dbKey = ctxDBKey[0]
	}
	return &BaseRepo{db: db, TableName: e.TableName(), IDField: e.IDField(), ctxDBKey: dbKey}
}

func (b *BaseRepo) GetDB(ctx context.Context) *gorm.DB {
	db, ok := ctx.Value(b.ctxDBKey).(*gorm.DB)
	if !ok {
		return b.db
	}
	return db
}
