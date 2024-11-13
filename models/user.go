package models

//user model

import (
	"time"
)

// User 用户数据结构
type User struct {
	ID          uint       `json:"id" gorm:"primary_key"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at" sql:"index"`
	Username    string     `gorm:"unique"`
	Password    string
	Phone       string `gorm:"type:varchar(255);"`
	Email       string `gorm:"type:varchar(255);"`
	Description string // 个人简介
}
