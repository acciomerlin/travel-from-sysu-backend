package models

import (
	"time"
)

// User 用户数据结构
type User struct {
	UserId        uint       `json:"uid" gorm:"primaryKey;autoIncrement;autoIncrementStart:100001"` // 用户ID，从100001开始递增
	CreatedAt     time.Time  `json:"created_at"`                                                    // 创建时间
	UpdatedAt     time.Time  `json:"updated_at"`                                                    // 更新时间
	DeletedAt     *time.Time `json:"deleted_at" sql:"index"`                                        // 删除时间（软删除支持）
	Username      string     `gorm:"unique;not null" json:"username"`                               // 用户名
	Password      string     `gorm:"not null" json:"password"`                                      // 密码
	Phone         string     `gorm:"type:varchar(50);" json:"phone"`                                // 手机号
	Email         string     `gorm:"type:varchar(100);" json:"email"`                               // 邮箱
	Description   string     `gorm:"type:longtext" json:"description"`                              // 个人简介
	Avatar        string     `gorm:"type:varchar(225);" json:"avatar"`                              // 用户头像
	Gender        *int       `json:"gender"`                                                        // 性别 (1: 男, 2: 女, 0: 未知)
	Status        *int       `json:"status"`                                                        // 状态 (0: 正常, 1: 禁用)
	UserCover     string     `gorm:"type:longtext;" json:"user_cover"`                              // 用户封面
	Birthday      string     `gorm:"type:varchar(50);" json:"birthday"`                             // 生日
	TrendCount    uint64     `gorm:"default:0" json:"trend_count"`                                  // 动态数量
	FollowerCount uint64     `gorm:"default:0" json:"follower_count"`                               // 关注人数
	FanCount      uint64     `gorm:"default:0" json:"fan_count"`                                    // 粉丝人数
	NoteCount     uint64     `gorm:"default:0" json:"note_count"`                                   // 发布帖子的数量
}
