package utils

import (
	"fmt"

	"github.com/go-playground/validator"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct validates a struct based on tags
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		// Return first validation error
		for _, err := range err.(validator.ValidationErrors) {
			return fmt.Errorf("field %s is %s", err.Field(), err.Tag())
		}
	}
	return nil
}
