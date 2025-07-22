package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/http/handler/middleware"
	"github.com/dgdraganov/user-api/internal/http/payload"
	"go.uber.org/zap"
)

var (
	Authenticate = "POST /api/auth"
	ListUsers    = "GET /api/users"
	//GetUser      = "GET /api/users/{id}"
	SaveFile = "POST /api/users/{id}/files"

	// Register     = "POST /api/register"
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
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	var payload payload.AuthRequest
	err := h.requestValidator.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		respond(w, Response{
			Message: "could not authenticate",
			Error:   badRequestErr,
		}, http.StatusBadRequest, requestId)
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
			resp.Error = oopsErr
		}

		respond(w, resp, httpCode, requestId)
		h.logs.Errorw("authentication failed",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	resp := map[string]string{
		"token": token,
	}
	respond(w, resp, http.StatusOK, requestId)
}

func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	var payload payload.UserListRequest
	err := h.requestValidator.DecodeAndValidateQueryParams(r, &payload)
	if err != nil {
		respond(w, Response{
			Message: "could not list users",
			Error:   badRequestErr,
		}, http.StatusBadRequest, requestId)
		h.logs.Errorw("failed to decode and validate query parameters",
			"error", err,
			"handler", ListUsers,
			"request_id", requestId)
		return
	}

	users, err := h.userSvc.ListUsers(r.Context(), payload.Page, payload.PageSize)
	if err != nil {
		respond(w, Response{
			Message: "could not list users",
			Error:   oopsErr,
		}, http.StatusInternalServerError, requestId)
		h.logs.Errorw("failed to list users",
			"error", err,
			"handler", ListUsers,
			"request_id", requestId)
		return
	}

	resp := map[string]interface{}{
		"users": users,
	}
	respond(w, resp, http.StatusOK, requestId)
}

func respond(w http.ResponseWriter, resp any, code int, requestId string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, oopsErr, http.StatusInternalServerError)
	}
}
