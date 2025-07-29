package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/dgdraganov/user-api/internal/http/payload"
	"github.com/dgdraganov/user-api/internal/service"
	"github.com/golang-jwt/jwt"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name RequestValidator . RequestValidator
type RequestValidator interface {
	DecodeAndValidateJSONPayload(r *http.Request, object any) error
	DecodeAndValidateQueryParams(r *http.Request, object payload.URLDecoder) error
}

//counterfeiter:generate -o fake -fake-name UserService . UserService
type UserService interface {
	Authenticate(ctx context.Context, msg service.AuthMessage) (string, error)
	ListUsers(ctx context.Context, page int, pageSize int) ([]service.UserRecord, error)
	ValidateToken(ctx context.Context, token string) (jwt.MapClaims, error)
	UploadUserFile(ctx context.Context, objectName string, file io.Reader, fileSize int64) error
	SaveFileMetadata(ctx context.Context, fileName, bucket, userID string) error
	GetUser(ctx context.Context, id string) (service.UserRecord, error)
	RegisterUser(ctx context.Context, msg service.RegisterMessage) error
	UpdateUser(ctx context.Context, msg service.UpdateUserMessage, userID string) error
	PublishEvent(ctx context.Context, routingKey string, payload interface{}) error
	DeleteUser(ctx context.Context, userID string) error
	ListUserFiles(ctx context.Context, resourceGUID string) ([]service.FileRecord, error)
	DeleteUserFiles(ctx context.Context, resourceGUID string) error
}
