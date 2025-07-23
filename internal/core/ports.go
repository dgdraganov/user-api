package core

import (
	"context"
	"io"

	"github.com/dgdraganov/user-api/internal/repository"
	tokenIssuer "github.com/dgdraganov/user-api/pkg/jwt"

	"github.com/golang-jwt/jwt"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name Repository . Repository
type Repository interface {
	GetUserByEmail(ctx context.Context, email string, user *repository.User) error
	ListUsersByPage(ctx context.Context, page int, pageSize int, users *[]repository.User) error
	SaveFileMetadata(ctx context.Context, fileMetadata repository.FileMetadata) error
	GetUserByID(ctx context.Context, id string, user *repository.User) error
	CreateUser(ctx context.Context, user repository.User) error
	UpdateUser(ctx context.Context, user repository.User) error
}

//counterfeiter:generate -o fake -fake-name JWTIssuer . JWTIssuer
type JWTIssuer interface {
	Generate(data tokenIssuer.TokenInfo) *jwt.Token
	Sign(token *jwt.Token) (string, error)
	Validate(token string) (jwt.MapClaims, error)
}

//counterfeiter:generate -o fake -fake-name BlobStorage . BlobStorage
type BlobStorage interface {
	UploadFile(ctx context.Context, bucketName, objectName string, file io.Reader, fileSize int64) error
}

//counterfeiter:generate -o fake -fake-name MessageBroker . MessageBroker
type MessageBroker interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}
