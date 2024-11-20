package models

//user model

import (
	"time"
)

// User 用户数据结构
type User struct {
	ID          uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt   time.Time  `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"column:updatedAt" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index;column:deletedAt" json:"deletedAt"`
	Username    string     `gorm:"unique;column:username" json:"username"`
	Password    string     `gorm:"column:password" json:"password"`
	Phone       string     `gorm:"type:varchar(255);column:phone" json:"phone"`
	Email       string     `gorm:"type:varchar(255);column:email" json:"email"`
	Description string     `gorm:"column:description" json:"description"` // 个人简介
}
