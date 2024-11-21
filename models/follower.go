package models

import "time"

type Follower struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	Uid       uint       `gorm:"not null" json:"uid"`
	Fid       uint       `gorm:"not null" json:"fid"`
	User      User       `gorm:"foreignKey:Uid;AssociationForeignKey:Uid"` // 关联到关注者的用户信息
	Followed  User       `gorm:"foreignKey:Fid;AssociationForeignKey:Uid"` // 关联到被关注者的用户信息
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}
