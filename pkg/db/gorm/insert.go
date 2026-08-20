package dbgorm

import (
	"context"

	"transcode-demo/pkg/db/entity"
)

type InsertRepo[E entity.IEntity, K any] struct {
	*BaseRepo
}

func NewInsertRepo[E entity.IEntity, K any](baseRepo *BaseRepo) *InsertRepo[E, K] {
	return &InsertRepo[E, K]{baseRepo}
}

func (b *InsertRepo[E, K]) Insert(ctx context.Context, e *E) error {
	return b.GetDB(ctx).WithContext(ctx).Create(e).Error
}
