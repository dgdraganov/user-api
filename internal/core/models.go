package core

import (
	"errors"

	"github.com/dgdraganov/user-api/internal/repository"
)

var (
	// error user already exists
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrUserNotFound            = errors.New("user not found")
	ErrIncorrectPassword error = errors.New("incorrect password")
)

type AuthMessage struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRecord struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
}

type RegisterMessage struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
	Password  string `json:"password"`
}

type UpdateUserMessage struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
}

func (um UpdateUserMessage) ToUser(existingUser repository.User) repository.User {
	if um.FirstName != "" {
		existingUser.FirstName = um.FirstName
	}
	if um.LastName != "" {
		existingUser.LastName = um.LastName
	}
	if um.Email != "" {
		existingUser.Email = um.Email
	}
	if um.Age > 0 {
		existingUser.Age = um.Age
	}
	return existingUser
}
