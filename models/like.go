package models

import "time"

// Like 点赞表结构
type Like struct {
	LikeID     uint      `gorm:"primaryKey;autoIncrement;" json:"like_id"` // 主键 ID
	Uid        uint      `gorm:"not null" json:"uid"`
	Nid        *uint     `gorm:"null" json:"nid" ` // 笔记 ID（外键，关联 Note 表的 NoteID）
	User       User      `gorm:"foreignKey:Uid;AssociationForeignKey:Uid"`
	Note       Note      `gorm:"constraint:OnDelete:CASCADE;foreignKey:Nid;AssociationForeignKey:NoteID"`
	CreateDate time.Time `gorm:"type:datetime" json:"create_date"`
	Cid        *uint     `gorm:"null" json:"cid"` // 笔记 ID（外键，关联 Comment 表的 CommentID）
	Comment    Comments  `gorm:"foreignKey:Cid;AssociationForeignKey:CommentID"`
}
