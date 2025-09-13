package utils

import (
	"fmt"

	"github.com/go-playground/validator"
)

var validate *validator.Validate

// 2. init() function runs automatically
/**
Runs automatically when the package is imported
Runs before main() function
Runs only once per package
Used for package initialization
*/
func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)

	if err != nil {
		// Return first validating error
		for _, err := range err.(validator.ValidationErrors) {
			return fmt.Errorf("field %s is %s", err.Field(), err.Tag())
		}
	}
	return nil
}
