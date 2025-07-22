package core

type AuthMessage struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRecord struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
}
