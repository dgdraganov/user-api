package core

import (
	"context"
	"errors"
	"fmt"

	tokenIssuer "github.com/dgdraganov/user-api/pkg/jwt"

	"github.com/dgdraganov/user-api/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var ErrIncorrectPassword error = errors.New("incorrect password")
var ErrUserNotFound error = errors.New("user not found")

// UserService is a struct that provides methods to interact with the Ethereum node and the database.
type UserService struct {
	logs      *zap.SugaredLogger
	repo      Repository
	jwtIssuer JWTIssuer
}

// NewUserService is a constructor function for the UserService type.
func NewUserService(logger *zap.SugaredLogger, repo Repository, jwt JWTIssuer) *UserService {
	return &UserService{
		logs:      logger,
		repo:      repo,
		jwtIssuer: jwt,
	}
}

// Authenticate checks the provided email and password against the database. If the credentials are valid, it generates a JWT token for the user.
func (f *UserService) Authenticate(ctx context.Context, msg AuthMessage) (string, error) {
	user, err := f.repo.GetUserFromDB(ctx, msg.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("get user from db: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(msg.Password)); err != nil {
		return "", ErrIncorrectPassword
	}

	tokenInfo := tokenIssuer.TokenInfo{
		Email:      user.Email,
		Subject:    user.ID,
		Expiration: 24,
	}
	token := f.jwtIssuer.Generate(tokenInfo)
	signed, err := f.jwtIssuer.Sign(token)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signed, nil
}
