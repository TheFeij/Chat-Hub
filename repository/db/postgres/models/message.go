package models

// Message represents a message in the chat server
type Message struct {
	ID     uint   `gorm:"column:id;primaryKey"`
	Author string `gorm:"column:author;not null"`
	Text   string `gorm:"column:tex;max:1024;min=1;not null"`
	User   User   `gorm:"foreignKey:username;references:author"`
}
