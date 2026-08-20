package dbgorm

import (
	"context"
	"fmt"

	"transcode-demo/pkg/db/entity"
)

type DeleteRepo[E entity.IEntity, K any] struct {
	*BaseRepo
}

func NewDeleteRepo[E entity.IEntity, K any](baseRepo *BaseRepo) *DeleteRepo[E, K] {
	return &DeleteRepo[E, K]{baseRepo}
}

func (b *DeleteRepo[E, K]) Delete(ctx context.Context, id K) error {
	var obj E
	return b.GetDB(ctx).WithContext(ctx).Delete(&obj, fmt.Sprintf("%s = ?", b.IDField), id).Error
}
