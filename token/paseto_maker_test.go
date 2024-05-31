package token

import (
	"Chat-Server/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// TestPasetoMaker tests PasetoMaker
func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32, util.ALL))
	require.NoError(t, err)
	require.NotEmpty(t, maker)

	t.Run("OK", func(t *testing.T) {
		username := util.RandomUsername()
		duration := 1 * time.Minute

		issuedAt := time.Now()
		expiredAt := issuedAt.Add(duration)

		token, payload, err := maker.CreateToken(username, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		require.NotEmpty(t, payload)

		returnedPayload, err := maker.VerifyToken(token)
		require.NoError(t, err)
		require.NotEmpty(t, returnedPayload)

		require.Equal(t, username, returnedPayload.Username)
		require.NotZero(t, returnedPayload.ID)
		require.WithinDuration(t, issuedAt, returnedPayload.IssuedAt, time.Second)
		require.WithinDuration(t, expiredAt, returnedPayload.ExpiredAt, time.Second)
	})
	t.Run("PasetoTokenExpired", func(t *testing.T) {
		username := util.RandomUsername()
		duration := -time.Minute

		token, payload, err := maker.CreateToken(username, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		require.NotEmpty(t, payload)

		returnedPayload, err := maker.VerifyToken(token)
		require.Error(t, err)
		require.Equal(t, err, ErrExpiredToken)
		require.Nil(t, returnedPayload)
	})
	t.Run("InvalidPasetoToken", func(t *testing.T) {
		payload, err := maker.VerifyToken("invalid token")
		require.Error(t, err)
		require.Equal(t, ErrInvalidToken, err)
		require.Nil(t, payload)
	})

}
