package models

import "time"

// Note 笔记数据结构
type Note struct {
	ID             uint      `gorm:"primaryKey" json:"id"`                                                  // 主键 ID
	NoteTitle      string    `json:"noteTitle"`                                                             // 笔记标题
	NoteContent    string    `json:"noteContent"`                                                           // 笔记内容
	NoteCount      int       `json:"noteCount"`                                                             // 计数
	NoteTagList    string    `json:"noteTagList"`                                                           // 笔记标签列表（字符串类型）
	NoteType       string    `json:"noteType"`                                                              // 笔记类型
	NoteURLs       string    `json:"noteURLs"`                                                              // 笔记相关 URL
	NoteCreatorID  uint      `gorm:"not null;index" json:"noteCreatorID"`                                   // 创建者 ID（外键）
	NoteCreator    User      `gorm:"foreignKey:NoteCreatorID;AssociationForeignKey:Uid" json:"noteCreator"` // 关联到用户表的外键
	NoteUpdateTime time.Time `json:"noteUpdateTime"`                                                        // 笔记更新时间
}
