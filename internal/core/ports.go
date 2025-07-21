package core

import (
	"context"

	"github.com/dgdraganov/user-api/internal/repository"
	tokenIssuer "github.com/dgdraganov/user-api/pkg/jwt"

	"github.com/golang-jwt/jwt"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name Repository . Repository
type Repository interface {
	GetUserFromDB(ctx context.Context, email string) (*repository.User, error)
}

//counterfeiter:generate -o fake -fake-name JWTIssuer . JWTIssuer
type JWTIssuer interface {
	Generate(data tokenIssuer.TokenInfo) *jwt.Token
	Sign(token *jwt.Token) (string, error)
	Validate(token string) (jwt.MapClaims, error)
}
