package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/dgdraganov/user-api/internal/db"
	"github.com/google/uuid"
)

var ErrUserNotFound error = errors.New("user not found")

// UserRepository is a type that is used to interact with the database for user-related operations.
type UserRepository struct {
	db Storage
}

// NewUserRepository is a constructor function for the UserRepository type.
func NewUserRepository(db Storage) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// MigrateTables migrates the given tables to the database.
func (r *UserRepository) MigrateTables(tables ...any) error {
	err := r.db.MigrateTable(tables...)
	if err != nil {
		return fmt.Errorf("migrate table(s): %w", err)
	}
	return err
}

// GetUserByEmail retrieves a user by their email address.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string, user *User) error {
	err := r.db.GetOneBy(ctx, "email", email, user)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get user from db: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string, user *User) error {
	err := r.db.GetOneBy(ctx, "id", id, user)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get user from db: %w", err)
	}

	return nil
}

// SeedUserTable seeds the user table with a default set of users if it is empty.
func (r *UserRepository) SeedUserTable(ctx context.Context) error {
	users := []User{
		{
			ID:           uuid.NewString(),
			FirstName:    "Alice",
			LastName:     "Cooper",
			Age:          30,
			Role:         RoleAdmin,
			Email:        "alice@example.com",
			PasswordHash: "$2a$10$7PrikY/17DYiRAA6JlaGl.yo26gwhTT53ESuovxGWvWJ4HhvGI/GK",
		},
		{
			ID:           uuid.NewString(),
			FirstName:    "Bob",
			LastName:     "Marley",
			Age:          35,
			Role:         RoleAdmin,
			Email:        "bob@example.com",
			PasswordHash: "$2a$10$SHWr22XIYjY3/nLI6QOSJezr5KAB2AUs740F8NahmhBNsPsKacL8u",
		},
	}

	err := r.db.SeedTable(ctx, &users)
	if err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	return nil
}

func (r *UserRepository) ListUsersByPage(ctx context.Context, page int, pageSize int, users *[]User) error {
	err := r.db.ListByPage(ctx, page, pageSize, users)
	if err != nil {
		return fmt.Errorf("list users by page: %w", err)
	}
	return nil
}

func (r *UserRepository) SaveFileMetadata(ctx context.Context, fileMetadata FileMetadata) error {
	err := r.db.InsertToTable(ctx, fileMetadata)
	if err != nil {
		return fmt.Errorf("save file metadata: %w", err)
	}
	return nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user User) error {
	err := r.db.InsertToTable(ctx, user)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	err := r.db.DeleteByID(ctx, userID, &User{})
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user User) error {
	err := r.db.UpdateTable(ctx, user)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
