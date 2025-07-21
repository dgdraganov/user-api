package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/http/handler/middleware"
	"github.com/dgdraganov/user-api/internal/http/payload"
	"go.uber.org/zap"
)

var (
	Authenticate = "POST /api/auth"
)

type UserHandler struct {
	logs             *zap.SugaredLogger
	requestValidator RequestValidator
	userSvc          UserService
}

func NewUserHandler(logger *zap.SugaredLogger, requestValidator RequestValidator, userService UserService) *UserHandler {
	return &UserHandler{
		logs:             logger,
		requestValidator: requestValidator,
		userSvc:          userService,
	}
}

func (h *UserHandler) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	reqIdCtx := r.Context().Value(middleware.RequestIDKey)
	if reqIdCtx != nil {
		requestId = reqIdCtx.(string)
	}

	var payload payload.AuthRequest
	err := h.requestValidator.DecodeJSONPayload(r, &payload)
	if err != nil || payload.Validate() != nil {
		h.respond(w, Response{
			Message: "Could not authenticate",
			Error:   fmt.Errorf("invalid request payload: %w", err).Error(),
		}, http.StatusBadRequest,
			requestId)
		h.logs.Errorw("failed to decode and validate request payload",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	token, err := h.userSvc.Authenticate(r.Context(), payload.ToMessage())
	if err != nil {
		resp := Response{
			Message: "Login failed",
		}
		httpCode := http.StatusInternalServerError
		if errors.Is(err, core.ErrUserNotFound) {
			httpCode = http.StatusUnauthorized
			resp.Error = err.Error()
		} else if errors.Is(err, core.ErrIncorrectPassword) {
			httpCode = http.StatusUnauthorized
			resp.Error = err.Error()
		} else {
			httpCode = http.StatusInternalServerError
			resp.Error = "unexpected error occurred"
		}

		h.respond(w, resp, httpCode, requestId)
		h.logs.Errorw("authentication failed",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	resp := map[string]string{
		"token": token,
	}
	h.respond(w, resp, http.StatusOK, requestId)
}

func (h *UserHandler) respond(w http.ResponseWriter, resp any, code int, requestId string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, oopsErr, http.StatusInternalServerError)
		h.logs.Errorw("failed to encode response",
			"error", err,
			"request_id", requestId)
	}
}
