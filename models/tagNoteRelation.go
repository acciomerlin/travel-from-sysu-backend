package models

import "time"

// TagNoteRelation 表示 tag_note_relation 表的模型
type TagNoteRelation struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NID        uint      `gorm:"not null;index" json:"nid"` // 笔记 ID（外键，关联 Note 表的 NoteID）
	TID        string    `gorm:"not null;index" json:"tid"` // 标签 ID（外键，关联 Tag 表的 ID）
	Note       Note      `gorm:"foreignKey:NID;references:NoteID"`
	Tag        Tag       `gorm:"foreignKey:TID;references:ID"`
	CreatorID  uint      `gorm:"not null" json:"creator_id"` // 创建者 ID，外键
	Creator    User      `gorm:"foreignKey:CreatorID;references:UserId"`
	CreateDate time.Time `gorm:"type:datetime" json:"create_date"`
}
