package payload

import (
	"fmt"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/jellydator/validation"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a AuthRequest) Validate() error {
	err := validation.ValidateStruct(&a,
		validation.Field(&a.Email, validation.Required),
		validation.Field(&a.Password, validation.Required),
	)
	if err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}

	return nil
}

func (a *AuthRequest) ToMessage() core.AuthMessage {
	return core.AuthMessage{
		Email:    a.Email,
		Password: a.Password,
	}
}
