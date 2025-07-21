package repository

import "context"

type Storage interface {
	MigrateTable(tbl ...any) error
	SeedTable(ctx context.Context, records any) error
	GetOneBy(ctx context.Context, field string, value any, dest any) error
}
