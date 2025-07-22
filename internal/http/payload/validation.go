package payload

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jellydator/validation"
)

type URLDecoder interface {
	DecodeFromURLValues(url.Values) error
}

type DecodeValidator struct{}

func (dv DecodeValidator) DecodeAndValidateJSONPayload(r *http.Request, object any) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	decoder.DisallowUnknownFields()
	err := decoder.Decode(object)
	if err != nil {
		return fmt.Errorf("decoding json payload: %w", err)
	}

	if err := dv.validatePayload(object); err != nil {
		return fmt.Errorf("validating query params: %w", err)
	}
	return nil
}

func (dv DecodeValidator) DecodeAndValidateQueryParams(r *http.Request, object URLDecoder) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parsing form: %w", err)
	}

	if err := object.DecodeFromURLValues(r.Form); err != nil {
		return fmt.Errorf("decoding query params: %w", err)
	}
	if err := dv.validatePayload(object); err != nil {
		return fmt.Errorf("validating query params: %w", err)
	}
	return nil
}

func (dv *DecodeValidator) validatePayload(object any) error {
	t, ok := object.(validation.Validatable)
	if !ok {
		return nil
	}

	if err := t.Validate(); err != nil {
		return fmt.Errorf("object validation: %w", err)
	}

	return nil
}
