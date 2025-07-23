package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/http/handler/middleware"
	"github.com/dgdraganov/user-api/internal/http/payload"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

var (
	Authenticate = "POST /api/auth"
	ListUsers    = "GET /api/users"
	GetUser      = "GET /api/users/{guid}"
	UploadFile   = "POST /api/users/upload"
	UserRegister = "POST /api/users"
	UserUpdate   = "PUT /api/users"
	UserDelete   = "DELETE /api/users"
)

type UserHandler struct {
	logs             *zap.SugaredLogger
	requestValidator RequestValidator
	coreSvc          CoreService
}

func NewUserHandler(logger *zap.SugaredLogger, requestValidator RequestValidator, coreService CoreService) *UserHandler {
	return &UserHandler{
		logs:             logger,
		requestValidator: requestValidator,
		coreSvc:          coreService,
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
		}, http.StatusBadRequest)
		h.logs.Errorw("failed to decode and validate request payload",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	token, err := h.coreSvc.Authenticate(r.Context(), payload.ToMessage())
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

		respond(w, resp, httpCode)
		h.logs.Errorw("authentication failed",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	resp := map[string]string{
		"token": token,
	}
	respond(w, resp, http.StatusOK)
}

func (h *UserHandler) HandleRegisterUser(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	var payload payload.RegisterRequest
	err := h.requestValidator.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		respond(w, Response{
			Message: couldNotRegister,
			Error:   badRequestErr,
		}, http.StatusBadRequest)
		h.logs.Errorw("failed to decode and validate request payload",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	err = h.coreSvc.RegisterUser(r.Context(), payload.ToMessage())
	if errors.Is(err, core.ErrUserAlreadyExists) {
		respond(w, Response{
			Message: couldNotRegister,
			Error:   "User with this email already exists",
		}, http.StatusConflict)
		h.logs.Errorw("user already exists",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}
	if err != nil {
		respond(w, Response{
			Message: couldNotRegister,
			Error:   oopsErr,
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to register user",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	// send event to RabbitMQ
	err = h.coreSvc.PublishEvent(r.Context(), "user.event.registered", payload.ToMap())
	if err != nil {
		h.logs.Errorw("failed to publish user registered event",
			"error", err,
			"handler", UserUpdate,
			"request_id", requestId)
	}

	respond(w, Response{
		Message: "User registered successfully!",
	}, http.StatusCreated)
}

func (h *UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	claims, err := h.authenticate(r)
	if err != nil {
		respond(w, Response{
			Message: "could not update user",
			Error:   err.Error(),
		}, http.StatusUnauthorized)
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		respond(w, Response{
			Message: couldNotUpdateUser,
			Error:   "Invalid user ID in token",
		}, http.StatusInternalServerError)
		h.logs.Errorw("invalid user ID in token", "handler", UploadFile, "request_id", requestId)
		return
	}

	var payload payload.UpdateUserRequest
	err = h.requestValidator.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		respond(w, Response{
			Message: "could not update user",
			Error:   badRequestErr,
		}, http.StatusBadRequest)
		h.logs.Errorw("failed to decode and validate request payload",
			"error", err,
			"handler", UserUpdate,
			"request_id", requestId)
		return
	}

	err = h.coreSvc.UpdateUser(r.Context(), payload.ToMessage(), userID)
	if err != nil {
		respond(w, Response{
			Message: couldNotUpdateUser,
			Error:   oopsErr,
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to update user",
			"error", err,
			"handler", UserUpdate,
			"request_id", requestId)
		return
	}

	// send event to RabbitMQ
	err = h.coreSvc.PublishEvent(r.Context(), "user.event.updated", payload.ToMap())
	if err != nil {
		h.logs.Errorw("failed to publish user updated event",
			"error", err,
			"handler", UserUpdate,
			"request_id", requestId)
	}

	resp := map[string]string{
		"message": "User updated successfully!",
	}
	respond(w, resp, http.StatusOK)
}

func (h *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	claims, err := h.authenticate(r)
	if err != nil {
		respond(w, Response{
			Message: couldNotDeleteUser,
			Error:   err.Error(),
		}, http.StatusUnauthorized)
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		respond(w, Response{
			Message: couldNotDeleteUser,
			Error:   "Invalid user ID in token",
		}, http.StatusInternalServerError)
		h.logs.Errorw("invalid user ID in token", "handler", UserDelete, "request_id", requestId)
		return
	}

	err = h.coreSvc.DeleteUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, core.ErrUserNotFound) {
			respond(w, Response{
				Message: couldNotDeleteUser,
				Error:   "No user found with the provided ID",
			}, http.StatusNotFound)
			return
		}
		respond(w, Response{
			Message: couldNotDeleteUser,
			Error:   oopsErr,
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to delete user",
			"error", err,
			"handler", UserDelete,
			"request_id", requestId)
		return
	}

	// send event to RabbitMQ
	err = h.coreSvc.PublishEvent(r.Context(), "user.event.deleted", map[string]string{"user_id": userID})
	if err != nil {
		h.logs.Errorw("failed to publish user deleted event",
			"error", err,
			"handler", UserDelete,
			"request_id", requestId)
	}

	resp := map[string]string{
		"message": "User deleted successfully!",
	}
	respond(w, resp, http.StatusOK)
}

func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	_, err := h.authenticate(r)
	if err != nil {
		respond(w, Response{
			Message: listUsersFailed,
			Error:   err.Error(),
		}, http.StatusUnauthorized)
		return
	}

	var payload payload.UserListRequest
	err = h.requestValidator.DecodeAndValidateQueryParams(r, &payload)
	if err != nil {
		respond(w, Response{
			Message: listUsersFailed,
			Error:   badRequestErr,
		}, http.StatusBadRequest)
		h.logs.Errorw("failed to decode and validate query parameters",
			"error", err,
			"handler", ListUsers,
			"request_id", requestId)
		return
	}

	users, err := h.coreSvc.ListUsers(r.Context(), payload.Page, payload.PageSize)
	if err != nil {
		respond(w, Response{
			Message: "could not list users",
			Error:   oopsErr,
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to list users",
			"error", err,
			"handler", ListUsers,
			"request_id", requestId)
		return
	}

	resp := map[string]interface{}{
		"users": users,
	}
	respond(w, resp, http.StatusOK)
}

func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	_, err := h.authenticate(r)
	if err != nil {
		respond(w, Response{
			Message: couldNotGetUser,
			Error:   err.Error(),
		}, http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	userID := pathParts[len(pathParts)-1]

	user, err := h.coreSvc.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, core.ErrUserNotFound) {
			respond(w, Response{
				Message: couldNotGetUser,
				Error:   "No user is found with the provided ID",
			}, http.StatusNotFound)
			return
		}
		respond(w, Response{
			Message: couldNotGetUser,
			Error:   oopsErr,
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to get user",
			"error", err,
			"handler", GetUser,
			"request_id", requestId,
			"user_id", userID)
		return
	}

	resp := map[string]interface{}{
		"user": user,
	}
	respond(w, resp, http.StatusOK)
}

func (h *UserHandler) authenticate(r *http.Request) (jwt.MapClaims, error) {

	authToken := r.Header.Get("AUTH_TOKEN")
	if authToken == "" {
		return nil, errors.New("AUTH_TOKEN header is required")
	}

	claims, err := h.coreSvc.ValidateToken(r.Context(), authToken)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}

func respond(w http.ResponseWriter, resp any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, oopsErr, http.StatusInternalServerError)
	}
}
