package models

import "time"

// Like 点赞表结构
type Like struct {
	LikeID     uint      `gorm:"primaryKey;autoIncrement;" json:"like_id"` // 主键 ID
	Uid        uint      `gorm:"not null" json:"uid"`
	Nid        uint      `gorm:"not null;index" json:"nid"` // 笔记 ID（外键，关联 Note 表的 NoteID）
	User       User      `gorm:"foreignKey:Uid;AssociationForeignKey:Uid"`
	Note       Note      `gorm:"foreignKey:Nid;AssociationForeignKey:NoteID"`
	CreateDate time.Time `gorm:"type:datetime" json:"create_date"`
}
