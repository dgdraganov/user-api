package handler

const (
	oopsErr         = "Oops! Something went wrong. Please try again later."
	badRequestErr   = "Invalid request parameters."
	uploadFailed    = "File upload failed."
	listUsersFailed = "Failed to list users."
)

type Response struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
