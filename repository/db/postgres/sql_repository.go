package postgres

import (
	"Chat-Server/repository"
	"Chat-Server/repository/db/postgres/models"
	"fmt"
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
			panic(fmt.Errorf("cannot connect to database"))
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
