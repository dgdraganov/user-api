package handler

import (
	"context"
	"net/http"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/http/payload"
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
}
