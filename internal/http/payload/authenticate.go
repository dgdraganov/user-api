package payload

import (
	"fmt"

	"github.com/dgdraganov/user-api/internal/service"
	"github.com/jellydator/validation"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a AuthRequest) Validate() error {
	err := validation.ValidateStruct(&a,
		validation.Field(&a.Email, validation.Required, validation.Match(regexEmail)),
		validation.Field(&a.Password, validation.Required),
	)
	if err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}

	return nil
}

func (a *AuthRequest) ToMessage() service.AuthMessage {
	return service.AuthMessage{
		Email:    a.Email,
		Password: a.Password,
	}
}
