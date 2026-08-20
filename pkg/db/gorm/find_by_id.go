package dbgorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"transcode-demo/internal/models"
	"transcode-demo/pkg/db/entity"
)

type FindByIDRepo[E entity.IEntity, K any] struct {
	*BaseRepo
}

func NewFindByIDRepo[E entity.IEntity, K any](baseRepo *BaseRepo) *FindByIDRepo[E, K] {
	return &FindByIDRepo[E, K]{baseRepo}
}

func (b *FindByIDRepo[E, K]) FindByID(ctx context.Context, id K) (*E, error) {
	obj := new(E)
	err := b.GetDB(ctx).WithContext(ctx).First(obj, fmt.Sprintf("%s = ?", b.IDField), id).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, models.ErrModelNotFound
	case err != nil:
		return nil, err
	}
	return obj, nil
}
