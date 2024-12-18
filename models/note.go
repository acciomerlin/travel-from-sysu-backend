package models

// Note 笔记数据结构
type Note struct {
	NoteID           uint    `gorm:"primaryKey;autoIncrement;autoIncrementStart:100001" json:"note_id"` // 主键 ID
	NoteTitle        string  `json:"note_title"`                                                        // 笔记标题
	NoteContent      string  `json:"note_content"`                                                      // 笔记内容
	ViewCount        uint    `json:"view_count"`                                                        // 浏览计数
	NoteTagList      string  `json:"note_tag_list"`                                                     // 笔记标签列表（字符串类型）
	NoteType         string  `json:"note_type"`                                                         // 笔记类型
	NoteURLs         string  `json:"note_URLs"`                                                         // 笔记相关 URL
	NoteCreatorID    uint    `gorm:"not null;index" json:"note_creator_id"`                             // 创建者 ID（外键）
	User             User    `gorm:"foreignKey:NoteCreatorID;AssociationForeignKey:NoteCreatorID"`
	NoteUpdateTime   int64   `json:"note_update_time"` // 笔记更新时间 (Unix 时间戳)
	LikeCounts       int     `json:"like_counts"`
	CollectCounts    uint    `json:"collect_counts"`
	CommentCounts    uint    `json:"comment_counts"`
	IsFindingBuddy   int     `json:"is_finding_buddy"`  // 是否是找旅伴帖子 (0: 否, 1: 是)
	BuddyDescription string  `json:"buddy_description"` // 找旅伴的需求描述
	Score            float64 `json:"score"`
}
