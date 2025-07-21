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

func (r *UserRepository) GetUserFromDB(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.GetOneBy(ctx, "email", email, &user)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user from db: %w", err)
	}

	return &user, nil
}

// SeedUserTable seeds the user table with a default set of users if it is empty.
func (r *UserRepository) SeedUserTable(ctx context.Context) error {

	users := []User{
		{
			ID:           uuid.NewString(),
			FirstName:    "Alice",
			LastName:     "Cooper",
			Age:          30,
			Email:        "alice@example.com",
			PasswordHash: "$2a$10$7PrikY/17DYiRAA6JlaGl.yo26gwhTT53ESuovxGWvWJ4HhvGI/GK",
		},
		{
			ID:           uuid.NewString(),
			FirstName:    "Bob",
			LastName:     "Marley",
			Age:          35,
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
