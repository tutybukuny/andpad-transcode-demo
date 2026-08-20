package dbgorm

import (
	"context"

	"transcode-demo/pkg/db/entity"
)

type BulkInsertRepo[E entity.IEntity, K any] struct {
	*BaseRepo
}

func NewBulkInsertRepo[E entity.IEntity, K any](baseRepo *BaseRepo) *BulkInsertRepo[E, K] {
	return &BulkInsertRepo[E, K]{baseRepo}
}

func (b *InsertRepo[E, K]) BulkInsert(ctx context.Context, es []E) ([]*E, error) {
	err := b.GetDB(ctx).WithContext(ctx).Create(&es).Error
	if err != nil {
		return nil, err
	}
	pointers := make([]*E, 0, len(es))
	for i := range es {
		pointers = append(pointers, &es[i])
	}
	return pointers, err
}
