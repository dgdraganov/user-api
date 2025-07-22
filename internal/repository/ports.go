package repository

import "context"

type Storage interface {
	MigrateTable(tbl ...any) error
	SeedTable(ctx context.Context, records any) error
	GetOneBy(ctx context.Context, field string, value any, dest any) error
	ListByPage(ctx context.Context, page, pageSize int, entity any) error
}
