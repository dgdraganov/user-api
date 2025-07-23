package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ErrNotFound = errors.New("record not found")

// MySQL is a struct that provides methods to interact with a MySQL database using GORM.
type MySQL struct {
	DB *gorm.DB
}

// NewMySQL is a constructor function that initializes a new MySQL instance.
func NewMySqlDB(dsn string) (*MySQL, error) {
	db, err := connectWithRetry(dsn, 10)
	if err != nil {
		return nil, fmt.Errorf("connect to database with retries: %w", err)
	}

	return &MySQL{
		DB: db,
	}, nil
}

// MigrateTable migrates the provided tables to the database schema.
func (u *MySQL) MigrateTable(tbl ...any) error {
	err := u.DB.AutoMigrate(tbl...)
	if err != nil {
		return fmt.Errorf("failed to migrate table: %w", err)
	}

	return nil
}

// InsertToTable inserts the provided records into the specified table.
func (u *MySQL) InsertToTable(ctx context.Context, records any) error {
	if err := u.DB.Create(records).Error; err != nil {
		return fmt.Errorf("insert to table: %w", err)
	}
	return nil
}

func (u *MySQL) UpdateTable(ctx context.Context, records any) error {
	if err := u.DB.Save(records).Error; err != nil {
		return fmt.Errorf("update table: %w", err)
	}
	return nil
}

// GetOneBy retrieves a single record from the specified table where the given column matches the provided value.
func (u *MySQL) GetOneBy(ctx context.Context, column string, value any, entity any) error {
	query := fmt.Sprintf("%s = ?", column)
	err := u.DB.Where(query, value).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("getting record by %q: %w", column, err)
	}
	return nil
}

// GetAllBy retrieves all records from the specified table where the given column matches the provided value.
func (u *MySQL) GetAllBy(ctx context.Context, column string, value any, entity any) error {
	tx := u.DB.Where(fmt.Sprintf("%s IN (?)", column), value).Find(entity)
	if tx.Error != nil {
		return fmt.Errorf("getting records by %q: %w", column, tx.Error)
	}
	return nil
}

// GetAll retrieves all records from the specified table and stores them in the provided entity object
func (u *MySQL) GetAll(ctx context.Context, entity any) error {
	tx := u.DB.Find(entity)
	if tx.Error != nil {
		return fmt.Errorf("getting all records: %w", tx.Error)
	}
	return nil
}

// DeleteByID deletes a record from the specified table by its ID.
// func (u *MySQL) DeleteByID(ctx context.Context, id string, entity any) error {
// 	tx := u.DB.Where("id = ?", id).Delete(entity)
// 	if tx.Error != nil {
// 		return fmt.Errorf("deleting record by ID: %w", tx.Error)
// 	}

// 	return nil
// }

// DeleteByID deletes a record from the specified table by its ID.
func (u *MySQL) DeleteBy(ctx context.Context, key string, value any, entity any) error {
	tx := u.DB.Where(fmt.Sprintf("%s = ?", key), value).Delete(entity)
	if tx.Error != nil {
		return fmt.Errorf("deleting record by %q: %w", key, tx.Error)
	}

	return nil
}

// SeedTable checks if the table is empty and seeds it with the provided records if it is.
func (u *MySQL) SeedTable(ctx context.Context, records any) error {

	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("records type must be pointer to a slice: %T", records)
	}

	slice := v.Elem()
	if slice.Len() == 0 {
		return nil
	}

	var count int64

	elemType := slice.Index(0).Interface()
	if err := u.DB.Model(elemType).Count(&count).Error; err != nil {
		return fmt.Errorf("get model count: %w", err)
	}

	if count > 0 {
		return nil
	}

	if err := u.DB.Create(records).Error; err != nil {
		return fmt.Errorf("insert to table: %w", err)
	}

	return nil
}

// ListByPage retrieves a paginated list of records from the specified table.
func (u *MySQL) ListByPage(ctx context.Context, page, pageSize int, entity any) error {
	if page < 1 || pageSize < 1 {
		return fmt.Errorf("page and pageSize must be greater than 0: (page=%d, pageSize=%d)", page, pageSize)
	}

	offset := (page - 1) * pageSize
	if err := u.DB.WithContext(ctx).Offset(offset).Limit(pageSize).Find(entity).Error; err != nil {
		return fmt.Errorf("paginating records: %w", err)
	}

	return nil
}

func connectWithRetry(dsn string, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			return db, nil
		}
		<-time.After(time.Second * time.Duration(i+1))
	}

	return nil, fmt.Errorf("connect to database after %d retries: %w", maxRetries, err)
}
