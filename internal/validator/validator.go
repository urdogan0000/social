package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate validates a struct
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errMsg string
			for i, err := range validationErrors {
				if i > 0 {
					errMsg += ", "
				}
				errMsg += fmt.Sprintf("%s: %s", err.Field(), getErrorMessage(err))
			}
			return fmt.Errorf("validation failed: %s", errMsg)
		}
		return err
	}
	return nil
}

// getErrorMessage returns a user-friendly error message
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

