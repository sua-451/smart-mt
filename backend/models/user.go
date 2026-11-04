package models

import (
	"time"
)

// AppUser maps to the app_users table (authentication / role management).
type AppUser struct {
	UserID       uint      `gorm:"primaryKey;column:user_id;autoIncrement"                 json:"user_id"`
	FullName     string    `gorm:"column:full_name;type:varchar(120);not null"             json:"full_name"`
	Role         string    `gorm:"column:role;type:varchar(20);not null"                  json:"role"` // Manager | Teknisi
	Username     string    `gorm:"column:username;type:varchar(60);not null;uniqueIndex"   json:"username"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null"        json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"                       json:"created_at"`
}

func (AppUser) TableName() string { return "app_users" }

// UserWithCount is a transient read-model populated via JOIN.
type UserWithCount struct {
	AppUser
	AssignedCount int `gorm:"column:assigned_count" json:"assigned_count"`
}
