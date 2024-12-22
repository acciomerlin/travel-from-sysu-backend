package models

import "time"

// Notification 表结构
type Notification struct {
	ID          uint       `gorm:"primaryKey;autoIncrement;" json:"id"`   // 主键 ID
	InitiatorID uint       `gorm:"not null" json:"initiator_id"`          // 发起人 ID
	RecipientID uint       `gorm:"not null" json:"recipient_id"`          // 通知对象 ID
	Type        string     `gorm:"type:varchar(20);not null" json:"type"` // 通知类型：关注/点赞/收藏
	InitiatedAt time.Time  `gorm:"not null" json:"initiated_at"`          // 通知时间
	IsRead      bool       `gorm:"not null;default:false" json:"is_read"` // 已读/未读状态
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at"` // 可选，软删除字段
}
