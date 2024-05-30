package models

import (
	"gorm.io/gorm"
	"time"
)

// Session represents a user session
type Session struct {
	ID           uint           `gorm:"column:id;primaryKey"`
	UserUsername string         `gorm:"column:username;not null"`
	RefreshToken string         `gorm:"column:refresh_token;not null"`
	UserAgent    string         `gorm:"column:user_agent;not null"`
	ClientIP     string         `gorm:"column:client_ip;not null"`
	IsBlocked    bool           `gorm:"column:is_blocked;default:false;not null"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null"`
	ExpiresAt    time.Time      `gorm:"column:expires_at;not null"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at"`
	User         User
}
