package models

import "time"

// Tag 表示 t_tag 表的模型
type Tag struct {
	ID           string    `gorm:"type:varchar(50);primaryKey" json:"id"`   // 主键 ID
	TName        string    `gorm:"type:varchar(50);unique" json:"t_name"`   // 名称（唯一约束）
	LikeCount    int64     `gorm:"type:bigint;default:0" json:"like_count"` // 有这个tag的笔记点赞数量之和
	CollectCount int64     `gorm:"type:bigint;default:0" json:"like_count"` // 有这个tag的笔记收藏数量之和
	UseCount     int64     `gorm:"type:bigint;default:0" json:"like_count"` // 有这个tag的笔记数量之和
	Creator      string    `gorm:"type:varchar(50)" json:"creator"`         // 创建者
	CreateDate   time.Time `gorm:"type:datetime" json:"create_date"`        // 创建时间
	UpdateDate   time.Time `gorm:"type:datetime" json:"update_date"`        // 更新时间
}
