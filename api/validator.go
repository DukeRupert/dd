package api

import (
	"github.com/go-playground/validator/v10"
)

// Custom validator
type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator{
	return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return NewValidationError(err.Error())
	}
	return nil
}