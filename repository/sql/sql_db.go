package sql

import "gorm.io/gorm"

// db holds singleton instance of the database
var db *gorm.DB

func Init(address string) {
	// TODO: load configurations

	// TODO: initialize database
}
