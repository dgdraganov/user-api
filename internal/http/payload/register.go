package payload

import (
	"fmt"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/jellydator/validation"
)

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (r RegisterRequest) Validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.FirstName, validation.Required, validation.Length(2, 50)),
		validation.Field(&r.LastName, validation.Required, validation.Length(2, 50)),
		validation.Field(&r.Age, validation.Required, validation.Min(18), validation.Max(200)),
		validation.Field(&r.Email, validation.Required, validation.Match(regexEmail)),
		validation.Field(&r.Password, validation.Required, validation.Length(3, 100)),
	)
	if err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}
	return nil
}

func (r *RegisterRequest) ToMessage() core.RegisterMessage {
	return core.RegisterMessage{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Email:     r.Email,
		Age:       r.Age,
		Password:  r.Password,
	}
}
