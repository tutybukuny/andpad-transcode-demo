package db

import (
	"context"

	"transcode-demo/pkg/db/entity"
)

type IInsert[E entity.IEntity, K any] interface {
	Insert(ctx context.Context, e *E) error
}

type IBulkInsert[E entity.IEntity, K any] interface {
	BulkInsert(ctx context.Context, es []E) ([]*E, error)
}

type IUpdate[E entity.IEntity, K any] interface {
	Update(ctx context.Context, e *E) error
}

type IDelete[E entity.IEntity, K any] interface {
	Delete(ctx context.Context, id K) error
}

type IFindByID[E entity.IEntity, K any] interface {
	FindByID(ctx context.Context, id K) (*E, error)
}
