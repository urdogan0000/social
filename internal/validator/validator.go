package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidationError represents a structured validation error
type ValidationError struct {
	Field   string
	Tag     string
	Param   string
	Message string
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var errMsg string
	for i, err := range ve {
		if i > 0 {
			errMsg += ", "
		}
		errMsg += fmt.Sprintf("%s: %s", err.Field, err.Message)
	}
	return fmt.Sprintf("validation failed: %s", errMsg)
}

func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errors ValidationErrors
			for _, err := range validationErrors {
				errors = append(errors, ValidationError{
					Field:   err.Field(),
					Tag:     err.Tag(),
					Param:   err.Param(),
					Message: getErrorMessage(err),
				})
			}
			return errors
		}
		return err
	}
	return nil
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s characters", err.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", err.Param())
	default:
		return fmt.Sprintf("failed validation for tag '%s'", err.Tag())
	}
}

