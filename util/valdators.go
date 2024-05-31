package util

import (
	"fmt"
	"regexp"
)

// ValidateUsername validates username
func ValidateUsername(username string) error {
	if len(username) < 4 {
		return fmt.Errorf("username must be at least 4 characters")
	}
	if len(username) > 64 {
		return fmt.Errorf("username must be at most 64 characters")
	}

	if match, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]*$", username); !match {
		return fmt.Errorf("username must contain only alphabets, digits and underscore. and must start with an alphabet")
	}

	return nil
}

// ValidatePassword validates password
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 64 {
		return fmt.Errorf("password must be at most 64 characters")
	}

	if match, _ := regexp.MatchString("^[a-zA-Z0-9_!@#$%&*^.]*$", password); !match {
		return fmt.Errorf("invalid character in password; only alphabets, digits, and the following special characters are allowed: _!@#$%%&*.^")
	}

	return nil
}
