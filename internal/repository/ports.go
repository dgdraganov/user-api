package repository

import (
	"context"
)

type Storage interface {
	MigrateTable(tbl ...any) error
	SeedTable(ctx context.Context, records any) error
	GetOneBy(ctx context.Context, field string, value any, dest any) error
	GetAllBy(ctx context.Context, column string, value any, entity any) error
	ListByPage(ctx context.Context, page, pageSize int, entity any) error
	InsertToTable(ctx context.Context, records any) error
	UpdateTable(ctx context.Context, records any) error
	DeleteBy(ctx context.Context, key string, value any, entity any) error
}
