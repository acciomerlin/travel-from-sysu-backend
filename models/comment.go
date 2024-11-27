package models

import "time"

type Comments struct {
	CommentId   uint      `json:"comment_id" gorm:"primaryKey;autoIncrement;autoIncrementStart:100001"` // 用户ID，从100001开始递增
	NoteId      uint      `json:"note_id" gorm:"not null;index;foreignKey:NoteId;references:NoteId"`
	CreatorId   uint      `json:"creator_id"`
	ParentId    uint      `json:"parent_id"`
	ReplyId     uint      `json:"reply_id"`
	ReplyUid    uint      `json:"reply_uid"`
	Level       int       `json:"level"`
	Content     string    `json:"content" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	CommentLike uint      `json:"comment_like"`
}
