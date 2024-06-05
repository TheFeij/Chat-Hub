package postgres

import (
	"Chat-Server/repository"
	"Chat-Server/repository/db/postgres/models"
	"github.com/rs/zerolog/log"
	driver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sync"
)

// PostgresRepository implements Repository
type PostgresRepository struct {
	db *gorm.DB
}

// ensure PostgresRepository implements Repository interface
var _ = (*PostgresRepository)(nil)

// postgres is the singleton instance of SQLDatabase
var postgresRepository PostgresRepository

// once is used to ensure the singleton instance is initialize once
var once sync.Once

// GetPostgresRepository returns a new PostgresRepository
func GetPostgresRepository(address string) *PostgresRepository {
	once.Do(func() {
		// get a db session
		db, err := gorm.Open(driver.Open(address))
		if err != nil {
			log.Fatal().Err(err).Msg("cannot connect to database")
		}

		// migrate models
		db.AutoMigrate(&models.User{})
		db.AutoMigrate(&models.Message{})
		db.AutoMigrate(&models.Session{})

		postgresRepository = PostgresRepository{
			db: db,
		}
	})

	return &postgresRepository
}

// AddMessage saves the input message to the postgres database
func (p *PostgresRepository) AddMessage(message *repository.Message) (*repository.Message, error) {
	// initialize a message model
	newMessage := models.Message{
		Text:   message.Text,
		Author: message.Author,
	}

	// save the message to the database
	if err := p.db.Create(&newMessage).Error; err != nil {
		return nil, err
	}

	return message, nil
}

// GetAllMessages retrieves all messages from the database
func (p *PostgresRepository) GetAllMessages() (messages []*repository.Message, err error) {
	err = p.db.
		Raw("SELECT * FROM messages ORDER BY messages.id ASC").
		Scan(&messages).Error

	return
}

// AddUser saves the input user into the postgres database
func (p *PostgresRepository) AddUser(user *repository.User) (*repository.User, error) {
	newUser := models.User{
		Username: user.Username,
		Password: user.Password,
	}

	if err := p.db.Create(&newUser).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves user by username from the postgres database
func (p *PostgresRepository) GetUser(username string) (user *repository.User, err error) {
	res := p.db.
		Model(models.User{}).
		Where("username = ?", username).
		Scan(&user)

	if res.RowsAffected == 0 {
		err = gorm.ErrRecordNotFound
	}

	return
}
