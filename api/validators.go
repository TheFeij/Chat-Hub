package api

import (
	"Chat-Server/util"
	"github.com/go-playground/validator/v10"
)

// ValidUsername gin validator for username
var ValidUsername validator.Func = func(fl validator.FieldLevel) bool {
	if username, ok := fl.Field().Interface().(string); ok {
		if err := util.ValidateUsername(username); err != nil {
			return false
		}
		return true
	}
	return false
}

// ValidPassword gin validator for password
var ValidPassword validator.Func = func(fl validator.FieldLevel) bool {
	if password, ok := fl.Field().Interface().(string); ok {
		if err := util.ValidatePassword(password); err != nil {
			return false
		}
		return true
	}
	return false
}
