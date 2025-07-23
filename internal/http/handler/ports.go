package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/http/payload"
	"github.com/golang-jwt/jwt"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name RequestValidator . RequestValidator
type RequestValidator interface {
	DecodeAndValidateJSONPayload(r *http.Request, object any) error
	DecodeAndValidateQueryParams(r *http.Request, object payload.URLDecoder) error
}

//counterfeiter:generate -o fake -fake-name CoreService . CoreService
type CoreService interface {
	Authenticate(ctx context.Context, msg core.AuthMessage) (string, error)
	ListUsers(ctx context.Context, page int, pageSize int) ([]core.UserRecord, error)
	ValidateToken(ctx context.Context, token string) (jwt.MapClaims, error)
	UploadUserFile(ctx context.Context, objectName string, file io.Reader, fileSize int64) error
	SaveFileMetadata(ctx context.Context, fileName, bucket, userID string) error
	GetUser(ctx context.Context, id string) (core.UserRecord, error)
	RegisterUser(ctx context.Context, msg core.RegisterMessage) error
	UpdateUser(ctx context.Context, msg core.UpdateUserMessage, userID string) error
	PublishEvent(ctx context.Context, routingKey string, payload interface{}) error
	DeleteUser(ctx context.Context, userID string) error
	ListUserFiles(ctx context.Context, resourceGUID string) ([]core.FileRecord, error)
	DeleteUserFiles(ctx context.Context, resourceGUID string) error
}
