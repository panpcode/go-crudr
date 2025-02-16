package structs

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("uuid4_or_empty", validateUUID4rEmpty)
}

func validateUUID4rEmpty(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	if id == "" {
		return true
	}
	_, err := uuid.Parse(id)
	return err == nil
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}
