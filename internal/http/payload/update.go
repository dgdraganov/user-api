package payload

import (
	"fmt"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/jellydator/validation"
)

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Email     string `json:"email"`
}

func (r UpdateUserRequest) Validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.FirstName, validation.Length(2, 50)),
		validation.Field(&r.LastName, validation.Length(2, 50)),
		validation.Field(&r.Age, validation.Min(18), validation.Max(200)),
		validation.Field(&r.Email, validation.Match(regexEmail)),
	)
	if err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}
	return nil
}

func (r UpdateUserRequest) ToMessage() core.UpdateUserMessage {
	return core.UpdateUserMessage{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Email:     r.Email,
		Age:       r.Age,
	}
}
