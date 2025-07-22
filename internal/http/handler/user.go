package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/http/handler/middleware"
	"github.com/dgdraganov/user-api/internal/http/payload"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

var (
	Authenticate = "POST /api/auth"
	ListUsers    = "GET /api/users"
	GetUser      = "GET /api/users/{id}"
	UploadFile   = "POST /api/users/upload"
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

func (h *UserHandler) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	claims, err := h.authenticate(r)
	if err != nil {
		respond(w, Response{
			Message: uploadFailed,
			Error:   err.Error(),
		}, http.StatusUnauthorized)
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		respond(w, Response{
			Message: uploadFailed,
			Error:   "Invalid user ID in token",
		}, http.StatusInternalServerError)
		h.logs.Errorw("invalid user ID in token", "handler", UploadFile, "request_id", requestId)
		return
	}

	// 20 MB max size
	r.ParseMultipartForm(20 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		respond(w, Response{
			Message: uploadFailed,
			Error:   fmt.Sprintf("Failed to retrieve file: %v", err),
		}, http.StatusBadRequest)
		h.logs.Errorw("failed to retrieve file", "handler", UploadFile, "request_id", requestId, "error", err)
		return
	}
	defer file.Close()

	err = h.coreSvc.SaveFileMetadata(r.Context(), handler.Filename, "user-files-bucket", userID)
	if err != nil {
		respond(w, Response{
			Message: uploadFailed,
			Error:   fmt.Sprintf("Failed to upload file %q", handler.Filename),
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to upload file",
			"error", err,
			"handler", UploadFile,
			"request_id", requestId)
		return
	}

	objectName := fmt.Sprintf("%s/%s", userID, handler.Filename)

	// todo: add content type
	// contentType := handler.Header.Get("Content-Type")
	err = h.coreSvc.UploadUserFile(r.Context(), objectName, file, handler.Size)
	if err != nil {
		respond(w, Response{
			Message: uploadFailed,
			Error:   oopsErr,
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to upload file",
			"error", err,
			"handler", UploadFile,
			"request_id", requestId)
		return
	}

	resp := map[string]string{
		"message": fmt.Sprintf("File %q uploaded successfully", handler.Filename),
	}
	respond(w, resp, http.StatusOK)
}

func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	// requestId := ""
	// if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
	// 	requestId = reqId
	// }

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
