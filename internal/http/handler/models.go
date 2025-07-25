package handler

const (
	oopsErr            = "Oops! Something went wrong. Please try again later."
	badRequestErr      = "Invalid request parameters."
	uploadFailed       = "File upload failed."
	listUsersFailed    = "Failed to list users."
	couldNotGetUser    = "Could not get user."
	couldNotRegister   = "Could not register user."
	couldNotUpdateUser = "Could not update user."
	couldNotDeleteUser = "Could not delete user."
)

type Response struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
