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

//counterfeiter:generate -o fake -fake-name UserService . UserService
type UserService interface {
	Authenticate(ctx context.Context, msg core.AuthMessage) (string, error)
	ListUsers(ctx context.Context, page int, pageSize int) ([]core.UserRecord, error)
	ValidateToken(ctx context.Context, token string) (jwt.MapClaims, error)
}

//counterfeiter:generate -o fake -fake-name FileService . FileService
type FileService interface {
	UploadUserFile(ctx context.Context, objectName string, file io.Reader, fileSize int64) error
	ValidateToken(ctx context.Context, token string) (jwt.MapClaims, error)
	SaveFileMetadata(ctx context.Context, fileName, bucket, userID string) error
}
