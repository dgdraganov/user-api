package handler

import (
	"context"
	"net/http"

	"github.com/dgdraganov/user-api/internal/core"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name RequestValidator . RequestValidator
type RequestValidator interface {
	DecodeJSONPayload(r *http.Request, object any) error
}

//counterfeiter:generate -o fake -fake-name UserService . UserService
type UserService interface {
	Authenticate(ctx context.Context, msg core.AuthMessage) (string, error)
}
