package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
)

// PublishNoteRequest 发布笔记请求参数
type PublishNoteRequest struct {
	NoteTitle     string   `json:"noteTitle" binding:"required"`
	NoteContent   string   `json:"noteContent" binding:"required"`
	NoteCount     int      `json:"noteCount"`
	NoteTagList   []string `json:"noteTagList"` // 使用数组类型
	NoteType      string   `json:"noteType"`
	NoteURLs      string   `json:"noteURLs"`
	NoteCreatorID uint     `json:"noteCreatorID"`
}

// PublishNoteResponse 笔记发布成功的返回信息
type PublishNoteResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// PublishNote 发布笔记接口
// @Summary 发布笔记接口
// @Description 用户通过提供笔记标题、内容等信息来发布一篇新的笔记
// @Tags 笔记相关接口
// @Accept application/json
// @Produce application/json
// @Param data body PublishNoteRequest true "发布笔记请求参数"
// @Success 200 {object} PublishNoteResponse "笔记发布成功响应信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /publishNote [post]
func PublishNote(ctx *gin.Context) {
	var req PublishNoteRequest

	// 绑定 JSON 数据到 PublishNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 将 tagList 数组转换为逗号分隔的字符串
	tagListStr := strings.Join(req.NoteTagList, ",")

	// 创建 Note 实例
	note := models.Note{
		NoteTitle:      req.NoteTitle,
		NoteContent:    req.NoteContent,
		NoteCount:      req.NoteCount,
		NoteTagList:    tagListStr, // 将数组转换为字符串存储
		NoteType:       req.NoteType,
		NoteURLs:       req.NoteURLs,
		NoteCreatorID:  req.NoteCreatorID,
		NoteUpdateTime: time.Now(),
	}

	// 保存到数据库
	if err := global.Db.Create(&note).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "笔记发布失败",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, PublishNoteResponse{
		Status: "笔记发布成功",
		Code:   200,
	})
}
