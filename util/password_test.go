package util

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestCheckPassword(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		password := RandomPassword()

		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword)

		err = CheckPassword(password, hashedPassword)
		require.NoError(t, err)
	})
	t.Run("WrongPassword", func(t *testing.T) {
		wrongPassword := RandomPassword()
		correctPassword := RandomPassword()

		hashedPassword, err := HashPassword(correctPassword)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword)

		err = CheckPassword(wrongPassword, hashedPassword)
		require.Error(t, err)
		require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
	})
	t.Run("HashTooShort", func(t *testing.T) {
		password := RandomPassword()
		err := CheckPassword(password, "short hash")
		require.Error(t, err)
		require.EqualError(t, err, bcrypt.ErrHashTooShort.Error())
	})
}
