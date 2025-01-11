package models

import "time"

// Collect 收藏表结构
type Collect struct {
	CollectID  uint      `gorm:"primaryKey;autoIncrement;" json:"collect_id"`
	Uid        uint      `gorm:"not null" json:"uid"`       // 收藏用户 ID
	Nid        *uint     `gorm:"not null;index" json:"nid"` // 笔记 ID（外键，关联 Note 表的 NoteID）
	User       User      `gorm:"foreignKey:Uid;references:UserId"`
	Note       Note      `gorm:"foreignKey:Nid;references:NoteID"`
	CreateDate time.Time `gorm:"type:datetime" json:"create_date"` // 收藏时间
}
