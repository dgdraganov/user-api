package handler

import (
	"fmt"
	"net/http"

	"github.com/dgdraganov/user-api/internal/http/handler/middleware"
	"go.uber.org/zap"
)

var (
	UploadFile = "POST /api/files/upload"

/*
   Get the files stored for a particular user ID
   Add a file to a user record referred by user ID
   Delete all files linked to a user ID
*/

// Register     = "POST /api/register"
)

type FileHandler struct {
	logs             *zap.SugaredLogger
	requestValidator RequestValidator
	fileSvc          FileService
}

func NewFileHandler(logger *zap.SugaredLogger, requestValidator RequestValidator, fileService FileService) *FileHandler {
	return &FileHandler{
		logs:             logger,
		requestValidator: requestValidator,
		fileSvc:          fileService,
	}
}

func (h *FileHandler) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	if reqId, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
		requestId = reqId
	}

	authToken := r.Header.Get("AUTH_TOKEN")
	if authToken == "" {
		respond(w, Response{
			Message: uploadFailed,
			Error:   "AUTH_TOKEN header is required",
		}, http.StatusUnauthorized,
			requestId)
		h.logs.Errorw("missing AUTH_TOKEN header", "handler", UploadFile, "request_id", requestId)
		return
	}

	claims, err := h.fileSvc.ValidateToken(r.Context(), authToken)
	if err != nil {
		respond(w, Response{
			Message: uploadFailed,
			Error:   "Invalid token",
		}, http.StatusUnauthorized,
			requestId)
		h.logs.Errorw("invalid token", "handler", UploadFile, "request_id", requestId, "error", err)
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		respond(w, Response{
			Message: uploadFailed,
			Error:   "Invalid user ID in token",
		}, http.StatusInternalServerError, requestId)
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
		}, http.StatusBadRequest, requestId)
		h.logs.Errorw("failed to retrieve file", "handler", UploadFile, "request_id", requestId, "error", err)
		return
	}
	defer file.Close()

	err = h.fileSvc.SaveFileMetadata(r.Context(), handler.Filename, "user-files-bucket", userID)
	if err != nil {
		respond(w, Response{
			Message: uploadFailed,
			Error:   fmt.Sprintf("Failed to upload file %q", handler.Filename),
		}, http.StatusInternalServerError, requestId)
		h.logs.Errorw("failed to upload file",
			"error", err,
			"handler", UploadFile,
			"request_id", requestId)
		return
	}

	objectName := fmt.Sprintf("%s/%s", userID, handler.Filename)

	// todo: add content type
	// contentType := handler.Header.Get("Content-Type")
	err = h.fileSvc.UploadUserFile(r.Context(), objectName, file, handler.Size)
	if err != nil {
		respond(w, Response{
			Message: uploadFailed,
			Error:   oopsErr,
		}, http.StatusInternalServerError, requestId)
		h.logs.Errorw("failed to upload file",
			"error", err,
			"handler", UploadFile,
			"request_id", requestId)
		return
	}

	resp := map[string]string{
		"message": fmt.Sprintf("File %q uploaded successfully", handler.Filename),
	}
	respond(w, resp, http.StatusOK, requestId)
}
