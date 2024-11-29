package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
)

// UpdateNoteRequest 更新笔记的请求参数
type UpdateNoteRequest struct {
	NoteID      uint   `json:"note_id" binding:"required"` // 笔记 ID
	NoteTitle   string `json:"note_title"`                 // 笔记标题
	NoteContent string `json:"note_content"`               // 笔记内容
	NoteTagList string `json:"note_tag_list"`              // 笔记标签列表
	NoteType    string `json:"note_type"`                  // 笔记类型
	NoteURLs    string `json:"note_URLs"`                  // 笔记相关 URL
}

// UpdateNoteResponse 更新笔记的响应
type UpdateNoteResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

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

// DeleteNoteRequest 删除笔记请求参数
type DeleteNoteRequest struct {
	NoteID uint `json:"note_id" binding:"required"` // 笔记的唯一标识符
}

// DeleteNoteResponse 删除笔记的返回信息
type DeleteNoteResponse struct {
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
		NoteUpdateTime: time.Now().Unix(), // 设置时间戳
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

// DeleteNote 删除笔记接口
// @Summary 删除笔记接口
// @Description 根据笔记 ID 删除指定的笔记
// @Tags 笔记相关接口
// @Accept application/json
// @Produce application/json
// @Param data body DeleteNoteRequest true "删除笔记请求参数"
// @Success 200 {object} DeleteNoteResponse "笔记删除成功响应信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "笔记不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /deleteNote [post]
func DeleteNote(ctx *gin.Context) {
	var req DeleteNoteRequest

	// 绑定 JSON 数据到 DeleteNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 尝试从数据库中删除该笔记
	if err := global.Db.Delete(&models.Note{}, req.NoteID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "删除笔记失败",
		})
		return
	}

	// 检查删除是否成功
	rowsAffected := global.Db.RowsAffected
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Status: "失败",
			Code:   404,
			Error:  "笔记不存在或已被删除",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, DeleteNoteResponse{
		Status: "笔记删除成功",
		Code:   200,
	})
}

// UpdateNote 更新笔记接口
// @Summary 更新笔记接口
// @Description 用户根据笔记 ID 更新笔记的详细信息
// @Tags 笔记相关接口
// @Accept application/json
// @Produce application/json
// @Param data body UpdateNoteRequest true "更新笔记请求参数"
// @Success 200 {object} UpdateNoteResponse "笔记更新成功响应信息"
// @Failure 400 {object} UpdateNoteResponse "请求参数错误"
// @Failure 404 {object} UpdateNoteResponse "笔记不存在"
// @Failure 500 {object} UpdateNoteResponse "服务器内部错误"
// @Router /updateNote [put]
func UpdateNote(ctx *gin.Context) {
	var req UpdateNoteRequest

	// 绑定 JSON 数据到 UpdateNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, UpdateNoteResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 根据 NoteID 查找笔记
	var note models.Note
	if err := global.Db.First(&note, "note_id = ?", req.NoteID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, UpdateNoteResponse{
			Status: "失败",
			Code:   404,
			Error:  "笔记不存在",
		})
		return
	}

	// 更新笔记字段
	if req.NoteTitle != "" {
		note.NoteTitle = req.NoteTitle
	}
	if req.NoteContent != "" {
		note.NoteContent = req.NoteContent
	}
	if req.NoteTagList != "" {
		note.NoteTagList = req.NoteTagList
	}
	if req.NoteType != "" {
		note.NoteType = req.NoteType
	}
	if req.NoteURLs != "" {
		note.NoteURLs = req.NoteURLs
	}
	note.NoteUpdateTime = time.Now().Unix() // 设置时间戳

	// 保存更新到数据库
	if err := global.Db.Save(&note).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, UpdateNoteResponse{
			Status: "失败",
			Code:   500,
			Error:  "笔记更新失败",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, UpdateNoteResponse{
		Status: "笔记更新成功",
		Code:   200,
	})
}

// GetFoNotes 获取用户关注用户的帖子，支持游标分页，使用时间戳
func GetFoNotes(ctx *gin.Context) {
	// 获取请求参数
	userID := ctx.Query("user_id")
	num := ctx.Query("num")
	cursor := ctx.Query("cursor") // 游标，用于分页（时间戳）

	// 参数校验
	if userID == "" || num == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"msg":     "参数缺失",
		})
		return
	}

	// 默认最大条数
	limit := 30
	if n, err := strconv.Atoi(num); err == nil && n > 0 && n < 30 {
		limit = n
	}

	// 查询用户关注的用户 ID
	var followers []models.Follower
	if err := global.Db.Where("uid = ?", userID).Find(&followers).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询关注信息失败",
		})
		return
	}

	// 获取关注的用户 ID 列表
	var followedUserIDs []uint
	for _, follower := range followers {
		followedUserIDs = append(followedUserIDs, follower.Fid)
	}

	// 构造查询条件
	query := global.Db.Where("note_creator_id IN ?", followedUserIDs)
	if cursor != "" {
		// 游标为时间戳（Unix 时间）
		if timestamp, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			query = query.Where("note_update_time < ?", timestamp) // 返回最近更新的帖子，所以 <
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"success": false,
				"msg":     "无效的游标参数",
			})
			return
		}
	}

	// 查询帖子数据
	var notes []models.Note
	if err := query.Order("note_update_time DESC").Limit(limit).Find(&notes).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询笔记失败",
		})
		return
	}

	// 构造返回结果
	var responseNotes []gin.H
	var nextCursor string
	for _, note := range notes {
		responseNotes = append(responseNotes, gin.H{
			"note_id":          note.NoteID,
			"note_title":       note.NoteTitle,
			"note_content":     note.NoteContent,
			"note_like":        note.NoteLike,
			"note_favorite":    note.NoteFavorite,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime, // 时间戳直接返回，前端要解析！
		})
	}

	// 设置下一个游标
	if len(notes) > 0 {
		nextCursor = strconv.FormatInt(notes[len(notes)-1].NoteUpdateTime, 10)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor, // 下次分页使用的游标
		},
	})
}
