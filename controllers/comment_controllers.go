package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
)

// GetNotesByCreatorIDResponse 获取笔记响应结构
type GetNotesByCreatorIDResponse struct {
	Status string        `json:"status"`
	Code   int           `json:"code"`
	Data   []models.Note `json:"data,omitempty"`
	Error  string        `json:"error,omitempty"`
}

// GetNoteResponse 获取笔记的响应结构
type GetNoteResponse struct {
	Status string       `json:"status"`          // 响应状态
	Code   int          `json:"code"`            // 响应代码
	Note   *models.Note `json:"note,omitempty"`  // 笔记数据
	Error  string       `json:"error,omitempty"` // 错误信息
}

// GetCommentRequest 获取评论的请求参数
type GetCommentRequest struct {
	CommentId string `json:"comment_id" binding:"required"` // 要获取的评论 ID
}

// GetCommentResponse 获取评论的响应
type GetCommentResponse struct {
	Status   string           `json:"status"`
	Code     int              `json:"code"`
	Comments *models.Comments `json:"comments,omitempty"` // 返回评论详情
	Error    string           `json:"error,omitempty"`
}

type DeleteCommentRequest struct {
	CommentId string `json:"comment_id" binding:"required"` // 要删除的评论 ID
}

// DeleteCommentResponse 删除评论的响应
type DeleteCommentResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// PublishCommentRequest 发布评论的请求参数
type PublishCommentRequest struct {
	NoteId    uint   `json:"note_id" binding:"required"`    // 关联的笔记 ID
	CreatorId uint   `json:"creator_id" binding:"required"` // 评论创建者 ID
	ParentId  uint   `json:"parent_id"`                     // 父评论 ID（如果是回复）
	ReplyId   uint   `json:"reply_id"`                      // 回复的评论 ID（如果是回复）
	ReplyUid  uint   `json:"reply_uid"`                     // 被回复的用户 ID
	Level     int    `json:"level" binding:"required"`      // 评论层级
	Content   string `json:"content" binding:"required"`    // 评论内容
}

// PublishCommentResponse 发布评论的响应
type PublishCommentResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// GetSecondLevelCommentsRequest 获取二级评论的请求参数
type GetSecondLevelCommentsRequest struct {
	CommentId string `json:"comment_id" binding:"required"` // 一级评论 ID
}

// GetSecondLevelCommentsResponse 获取二级评论的响应
type GetSecondLevelCommentsResponse struct {
	Status   string            `json:"status"`
	Code     int               `json:"code"`
	Comments []models.Comments `json:"comments,omitempty"` // 返回评论数组
	Error    string            `json:"error,omitempty"`
}

// PublishComment 发布评论接口
// @Summary 发布评论接口
// @Description 用户发布评论
// @Tags 评论相关接口
// @Accept application/json
// @Produce application/json
// @Param data body PublishCommentRequest true "发布评论请求参数"
// @Success 200 {object} PublishCommentResponse "评论发布成功响应信息"
// @Failure 400 {object} PublishCommentResponse "请求参数错误"
// @Failure 500 {object} PublishCommentResponse "服务器内部错误"
// @Router /publishComment [post]
func PublishComment(ctx *gin.Context) {
	var req PublishCommentRequest

	// 绑定 JSON 数据到 PublishCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, PublishCommentResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 创建评论实例
	comment := models.Comments{
		NoteId:    req.NoteId,
		CreatorId: req.CreatorId,
		ParentId:  req.ParentId,
		ReplyId:   req.ReplyId,
		ReplyUid:  req.ReplyUid,
		Level:     req.Level,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	// 保存到数据库
	if err := global.Db.Create(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, PublishCommentResponse{
			Status: "失败",
			Code:   500,
			Error:  "评论发布失败",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, PublishCommentResponse{
		Status: "评论发布成功",
		Code:   200,
	})
}

// DeleteComment 删除评论接口
// @Summary 删除评论接口
// @Description 根据评论 ID 删除指定的评论
// @Tags 评论相关接口
// @Accept application/json
// @Produce application/json
// @Param data body DeleteCommentRequest true "删除评论请求参数"
// @Success 200 {object} DeleteCommentResponse "评论删除成功响应信息"
// @Failure 400 {object} DeleteCommentResponse "请求参数错误"
// @Failure 404 {object} DeleteCommentResponse "评论不存在"
// @Failure 500 {object} DeleteCommentResponse "服务器内部错误"
// @Router /deleteComment [post]
func DeleteComment(ctx *gin.Context) {
	var req DeleteCommentRequest

	// 绑定 JSON 数据到 DeleteCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, DeleteCommentResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 尝试从数据库中删除评论
	result := global.Db.Delete(&models.Comments{}, "comment_id = ?", req.CommentId)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, DeleteCommentResponse{
			Status: "失败",
			Code:   500,
			Error:  "删除评论失败",
		})
		return
	}

	// 检查是否有记录被删除
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, DeleteCommentResponse{
			Status: "失败",
			Code:   404,
			Error:  "评论不存在或已被删除",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, DeleteCommentResponse{
		Status: "评论删除成功",
		Code:   200,
	})
}

// GetCommentById 获取评论接口
// @Summary 获取评论接口
// @Description 根据评论 ID 获取当前评论的详细信息
// @Tags 评论相关接口
// @Accept application/json
// @Produce application/json
// @Param comment_id query string true "评论 ID"
// @Success 200 {object} GetCommentResponse "评论获取成功响应信息"
// @Failure 400 {object} GetCommentResponse "请求参数错误"
// @Failure 404 {object} GetCommentResponse "评论不存在"
// @Failure 500 {object} GetCommentResponse "服务器内部错误"
// @Router /getCommentById [get]
func GetCommentById(ctx *gin.Context) {
	// 获取查询参数
	commentId := ctx.Query("comment_id")
	if commentId == "" {
		ctx.JSON(http.StatusBadRequest, GetCommentResponse{
			Status: "失败",
			Code:   400,
			Error:  "评论 ID 不能为空",
		})
		return
	}

	// 查询数据库
	var comment models.Comments
	if err := global.Db.First(&comment, "comment_id = ?", commentId).Error; err != nil {
		ctx.JSON(http.StatusNotFound, GetCommentResponse{
			Status: "失败",
			Code:   404,
			Error:  "评论不存在",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, GetCommentResponse{
		Status:   "成功",
		Code:     200,
		Comments: &comment,
	})
}

// GetFirstLevelCommentsByNoteId 获取一级评论的接口
// @Summary 获取一级评论
// @Description 根据笔记 ID 获取所有一级评论
// @Tags 评论相关接口
// @Accept application/json
// @Produce application/json
// @Param note_id query string true "笔记的唯一标识符"
// @Success 200 {object} gin.H "成功返回一级评论"
// @Failure 400 {object} gin.H "请求参数错误"
// @Failure 404 {object} gin.H "未找到评论"
// @Failure 500 {object} gin.H "服务器错误"
// @Router /api/comment/getFirstLevelCommentByNoteId [get]
func GetFirstLevelCommentsByNoteId(ctx *gin.Context) {
	// 获取查询参数 note_id
	noteID := ctx.Query("note_id")
	if noteID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "笔记 ID 不能为空",
		})
		return
	}

	// 查询数据库中的一级评论（parentId 为空或为零的评论）
	var comments []models.Comments
	if err := global.Db.Where("note_id = ? AND (parent_id IS NULL OR parent_id = '' AND level = 1)", noteID).Find(&comments).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "获取评论失败",
		})
		return
	}

	// 检查是否有评论
	if len(comments) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "失败",
			"code":   404,
			"error":  "当前笔记没有评论",
		})
		return
	}

	// 成功响应：返回评论列表
	ctx.JSON(http.StatusOK, gin.H{
		"status":   "成功",
		"code":     200,
		"comments": comments, // 返回评论数组
	})
}

// GetSecondLevelCommentsByParentId 获取二级评论接口
// @Summary 获取二级评论
// @Description 根据一级评论 ID 获取所有二级评论
// @Tags 评论相关接口
// @Accept application/json
// @Produce application/json
// @Param data body GetSecondLevelCommentsRequest true "一级评论 ID"
// @Success 200 {object} GetSecondLevelCommentsResponse "成功返回二级评论列表"
// @Failure 400 {object} GetSecondLevelCommentsResponse "请求参数错误"
// @Failure 404 {object} GetSecondLevelCommentsResponse "未找到评论"
// @Failure 500 {object} GetSecondLevelCommentsResponse "服务器错误"
// @Router /api/comment/getSecondLevelComments [post]
func GetSecondLevelCommentsByParentId(ctx *gin.Context) {
	// 获取查询参数 note_id
	commentID := ctx.Query("comment_id")
	if commentID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "笔记 ID 不能为空",
		})
		return
	}

	// 查询数据库中的二级评论
	var comments []models.Comments
	if err := global.Db.Where("parent_id = ?", commentID).Find(&comments).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, GetSecondLevelCommentsResponse{
			Status: "失败",
			Code:   500,
			Error:  "获取评论失败",
		})
		return
	}

	// 检查是否有评论
	if len(comments) == 0 {
		ctx.JSON(http.StatusNotFound, GetSecondLevelCommentsResponse{
			Status: "无回复",
			Code:   404,
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, GetSecondLevelCommentsResponse{
		Status:   "成功",
		Code:     200,
		Comments: comments,
	})
}

// GetNoteByID 根据笔记 ID 获取笔记信息
// @Summary 根据笔记 ID 获取笔记
// @Description 根据笔记 ID 获取笔记详细信息
// @Tags 笔记相关接口
// @Accept application/json
// @Produce application/json
// @Param note_id path uint true "笔记 ID"
// @Success 200 {object} GetNoteResponse "笔记获取成功响应信息"
// @Failure 400 {object} GetNoteResponse "请求参数错误"
// @Failure 404 {object} GetNoteResponse "笔记不存在"
// @Router /getNote/{note_id} [get]
func GetNoteByID(ctx *gin.Context) {
	// 从路径参数获取 note_id
	noteID := ctx.Param("note_id")
	if noteID == "" {
		ctx.JSON(http.StatusBadRequest, GetNoteResponse{
			Status: "失败",
			Code:   400,
			Error:  "笔记id不得为空",
		})
		return
	}

	// 查询数据库中是否存在该笔记
	var note models.Note
	if err := global.Db.First(&note, "note_id = ?", noteID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, GetNoteResponse{
			Status: "失败",
			Code:   404,
			Error:  "笔记不存在",
		})
		return
	}

	// 返回笔记数据
	ctx.JSON(http.StatusOK, GetNoteResponse{
		Status: "成功",
		Code:   200,
		Note:   &note,
	})
}

// GetNotesByCreatorID 根据创建者 ID 获取笔记
// @Summary 根据创建者 ID 获取笔记
// @Description 根据创建者 ID 获取该用户创建的所有笔记
// @Tags 笔记相关接口
// @Accept application/json
// @Produce application/json
// @Param creator_id query string true "创建者 ID"
// @Success 200 {object} GetNotesByCreatorIDResponse "成功返回笔记数组"
// @Failure 400 {object} GetNotesByCreatorIDResponse "请求参数错误"
// @Failure 404 {object} GetNotesByCreatorIDResponse "未找到相关笔记"
// @Failure 500 {object} GetNotesByCreatorIDResponse "服务器内部错误"
// @Router /api/note/getNotesByCreatorId [get]
func GetNotesByCreatorID(ctx *gin.Context) {
	// 获取创建者 ID
	creatorID := ctx.Query("creator_id")
	if creatorID == "" {
		ctx.JSON(http.StatusBadRequest, GetNotesByCreatorIDResponse{
			Status: "失败",
			Code:   400,
			Error:  "创建者 ID 不能为空",
		})
		return
	}

	// 查询数据库中的笔记
	var notes []models.Note
	if err := global.Db.Where("note_creator_id = ?", creatorID).Find(&notes).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, GetNotesByCreatorIDResponse{
			Status: "失败",
			Code:   500,
			Error:  "查询笔记失败",
		})
		return
	}

	// 检查是否有笔记
	if len(notes) == 0 {
		ctx.JSON(http.StatusNotFound, GetNotesByCreatorIDResponse{
			Status: "失败",
			Code:   404,
			Error:  "未找到相关笔记",
		})
		return
	}

	// 成功响应：返回笔记列表
	ctx.JSON(http.StatusOK, GetNotesByCreatorIDResponse{
		Status: "成功",
		Code:   200,
		Data:   notes,
	})
}
