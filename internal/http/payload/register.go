package payload

import (
	"fmt"
	"regexp"

	"github.com/dgdraganov/user-api/internal/service"
	"github.com/jellydator/validation"
)

const (
	emailValidationRegex = `^(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`
	nameValidationRegex  = `^[A-Z][a-z]{1,}$`
)

var (
	regexEmail = regexp.MustCompile(emailValidationRegex)
	regexName  = regexp.MustCompile(nameValidationRegex)
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
		validation.Field(&r.FirstName, validation.Required, validation.Length(2, 50), validation.Match(regexName)),
		validation.Field(&r.LastName, validation.Required, validation.Length(2, 50), validation.Match(regexName)),
		validation.Field(&r.Age, validation.Required, validation.Min(18), validation.Max(200)),
		validation.Field(&r.Email, validation.Required, validation.Match(regexEmail)),
		validation.Field(&r.Password, validation.Required, validation.Length(3, 100)),
	)
	if err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}
	return nil
}

func (r RegisterRequest) ToMessage() service.RegisterMessage {
	return service.RegisterMessage{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Email:     r.Email,
		Age:       r.Age,
		Password:  r.Password,
	}
}

func (r RegisterRequest) ToMap() map[string]any {
	res := make(map[string]any)
	res["first_name"] = r.FirstName
	res["last_name"] = r.LastName
	res["email"] = r.Email
	res["age"] = r.Age
	return res
}
