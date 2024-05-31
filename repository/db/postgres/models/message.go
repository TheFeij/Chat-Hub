package models

// Message represents a message in the chat server
type Message struct {
	ID     uint   `gorm:"column:id;primaryKey"`
	Author string `gorm:"column:author;not null"`
	Text   string `gorm:"column:text;not null"`
	User   User   `gorm:"foreignKey:Author;references:Username"`
}
