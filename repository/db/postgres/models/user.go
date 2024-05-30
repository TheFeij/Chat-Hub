package models

// User represents a user in the database
type User struct {
	Username string `gorm:"column:username;primaryKey"`
	Password string `gorm:"column:password;not null"`
}
