package payload

import (
	"fmt"
	"math"
	"net/url"
	"strconv"

	"github.com/jellydator/validation"
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
