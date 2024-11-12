package models

//user model

import (
	"time"
)

// User 用户数据结构
type User struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
	Username  string     `gorm:"unique"`
	Password  string
}

// UserRegisterRequest 注册请求参数
type UserRegisterRequest struct {
	Username string `json:"username" example:"user123" binding:"required"`
	Password string `json:"password" example:"password123" binding:"required"`
}
