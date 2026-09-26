package handler

import "github.com/go-playground/validator/v10"

// RequestValidator adapts go-playground/validator to Echo's Validator interface.
// It is used only for HTTP-framework-level structural validation of request DTOs
// (e.g. rejecting path-separator characters); business rules stay in domain/.
type RequestValidator struct {
	validate *validator.Validate
}

func NewRequestValidator() *RequestValidator {
	return &RequestValidator{validate: validator.New()}
}

func (v *RequestValidator) Validate(i any) error {
	return v.validate.Struct(i)
}
