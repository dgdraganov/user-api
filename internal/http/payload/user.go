package payload

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/jellydator/validation"
)

const (
	emailValidationRegex = `^(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`
)

var (
	regexEmail = regexp.MustCompile(emailValidationRegex)
)

type UserListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (r UserListRequest) Validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.Page, validation.Required, validation.Min(1)),
		validation.Field(&r.PageSize, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return fmt.Errorf("validate struct: %w", err)
	}

	if int32(r.PageSize) > math.MaxInt32/int32(r.Page) {
		return fmt.Errorf("page %d and page_size %d would cause integer overflow", r.Page, r.PageSize)
	}

	return nil
}

func (r *UserListRequest) DecodeFromURLValues(values url.Values) error {
	pageStr := values.Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fmt.Errorf("parse page value: %w", err)
	}

	pageSizeStr := values.Get("page_size")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return fmt.Errorf("parse page_size value: %w", err)
	}
	r.Page = page
	r.PageSize = pageSize
	return nil
}

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
