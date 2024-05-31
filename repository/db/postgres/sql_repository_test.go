package postgres

import (
	"Chat-Server/repository"
	"Chat-Server/util"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

// adds a random user to the postgres database
func addRandomUser(t *testing.T) *repository.User {
	password := util.RandomPassword()

	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user := &repository.User{
		Username: util.RandomUsername(),
		Password: hashedPassword,
	}

	res, err := postgresRepository.AddUser(user)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, user.Username, res.Username)
	require.Equal(t, user.Password, res.Password)

	return res
}

// TestAddUser tests AddUser method of the PostgresRepository
func TestPostgresRepository_AddUser(t *testing.T) {
	defer cleanupDatabase()

	var randomUser *repository.User

	t.Run("OK", func(t *testing.T) {
		randomUser = addRandomUser(t)
	})
	t.Run("DuplicateUsername", func(t *testing.T) {
		res, err := postgresRepository.AddUser(randomUser)
		require.Nil(t, res)
		require.Error(t, err)

		// convert the error to pgError
		var pgError *pgconn.PgError
		ok := errors.As(err, &pgError)
		require.True(t, ok)
		require.Equal(t, "users_pkey", pgError.ConstraintName)
	})

}

// TestAddUser tests GetUser method of the PostgresRepository
func TestPostgresRepository_GetUser(t *testing.T) {
	defer cleanupDatabase()

	randomUser := addRandomUser(t)

	t.Run("OK", func(t *testing.T) {
		res, err := postgresRepository.GetUser(randomUser.Username)
		require.NoError(t, err)
		require.NotEmpty(t, res)

		require.Equal(t, randomUser.Username, res.Username)
		require.Equal(t, randomUser.Password, res.Password)
	})
	t.Run("NotFound", func(t *testing.T) {
		res, err := postgresRepository.GetUser("non existing username")
		require.Error(t, err)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		require.Nil(t, res)
	})
}

// addRandomMessage creates a random user as author and creates a message
// for that author and then returns the random user and its message
func addRandomMessage(t *testing.T, author string) *repository.Message {
	randomText := util.RandomText()

	message := &repository.Message{
		Author: author,
		Text:   randomText,
	}

	res, err := postgresRepository.AddMessage(message)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, author, res.Author)
	require.Equal(t, message.Text, res.Text)

	return message
}

// TestPostgresRepository_AddMessage tests AddMessage method of PostgresRepository
func TestPostgresRepository_AddMessage(t *testing.T) {
	defer cleanupDatabase()

	randomUser := addRandomUser(t)

	t.Run("OK", func(t *testing.T) {
		addRandomMessage(t, randomUser.Username)
	})
	t.Run("AuthorNotFound", func(t *testing.T) {
		message := &repository.Message{
			Author: "non existing author",
			Text:   util.RandomText(),
		}

		res, err := postgresRepository.AddMessage(message)
		require.Error(t, err)
		require.Nil(t, res)

		var pgError *pgconn.PgError
		ok := errors.As(err, &pgError)
		require.True(t, ok)
		require.Equal(t, "fk_messages_user", pgError.ConstraintName)
	})
}

// TestPostgresRepository_GetAllMessages tests GetAllMessages method of PostgresRepository
func TestPostgresRepository_GetAllMessages(t *testing.T) {
	defer cleanupDatabase()

	// fill the test database with some random data
	randomUser1 := addRandomUser(t)
	randomUser2 := addRandomUser(t)

	messages1 := make([]*repository.Message, 5)
	messages2 := make([]*repository.Message, 5)

	for i := 0; i < 5; i++ {
		var message *repository.Message
		message = addRandomMessage(t, randomUser1.Username)
		messages1[i] = message
	}

	for i := 0; i < 5; i++ {
		var message *repository.Message
		message = addRandomMessage(t, randomUser2.Username)
		messages2[i] = message
	}

	res, err := postgresRepository.GetAllMessages()
	require.NoError(t, err)
	require.NotEmpty(t, res)
	for _, message := range res {
		fmt.Println(*message)
	}
	require.Equal(t, 10, len(res))

	for i := 0; i < 5; i++ {
		require.Equal(t, messages1[i].Author, res[i].Author)
		require.Equal(t, messages1[i].Text, res[i].Text)
	}

	for i := 0; i < 5; i++ {
		require.Equal(t, messages2[i].Author, res[i+5].Author)
		require.Equal(t, messages2[i].Text, res[i+5].Text)
	}
}
