package models

import "time"

// 笔记数据结构

type Note struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	NoteTitle      string    `json:"noteTitle"`
	NoteContent    string    `json:"noteContent"`
	NoteCount      int       `json:"noteCount"`
	NoteTagList    string    `json:"noteTagList"` // 这里是字符串类型
	NoteType       string    `json:"noteType"`
	NoteURLs       string    `json:"noteURLs"`
	NoteCreatorID  int       `json:"noteCreatorID"`
	NoteUpdateTime time.Time `json:"noteUpdateTime"`
}
