package handler

import (
	"fmt"
	"net/http"

	"github.com/dgdraganov/user-api/internal/http/handler/middleware"
)

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
