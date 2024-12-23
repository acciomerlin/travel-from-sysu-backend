package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
	"travel-from-sysu-backend/utils"
)

// LikeOrCollectRequest 请求结构
type LikeOrCollectRequest struct {
	NoteID uint `json:"note_id" binding:"required"`
	Uid    uint `json:"uid" binding:"required"`
}

// LikeOrCollectResponse 响应结构
type LikeOrCollectResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// Like 点赞接口
func Like(ctx *gin.Context) {
	var req LikeOrCollectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 检查用户是否已经点赞过
	var existingLike models.Like
	if err := global.Db.Where("uid = ? AND nid = ?", req.Uid, req.NoteID).First(&existingLike).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  "您已点赞过此笔记",
		})
		return
	}

	// 添加 Like 表记录
	like := models.Like{
		Uid:        req.Uid,
		Nid:        req.NoteID,
		CreateDate: time.Now(),
	}
	if err := global.Db.Create(&like).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, LikeOrCollectResponse{
			Status: "失败",
			Code:   500,
			Error:  "点赞失败：" + err.Error(),
		})
		return
	}

	// 更新 Note 的 like_count
	if err := global.Db.Model(&models.Note{}).
		Where("note_id = ?", req.NoteID).
		Update("like_counts", gorm.Expr("like_counts + ?", 1)).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, LikeOrCollectResponse{
			Status: "失败",
			Code:   500,
			Error:  "更新 Note 点赞数失败",
		})
		return
	}

	// 更新相关 Tag 的 like_count
	var note models.Note
	if err := global.Db.First(&note, req.NoteID).Error; err == nil {
		tags := strings.Split(note.NoteTagList, ",")
		global.Db.Model(&models.Tag{}).
			Where("t_name IN ?", tags).
			Update("like_count", gorm.Expr("like_count + ?", 1))
	}

	// 添加通知记录
	if err := AddNotificationAndUpdateUnreadCount(req.Uid, note.NoteCreatorID, "like"); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "通知记录创建失败：" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, LikeOrCollectResponse{
		Status: "点赞成功",
		Code:   200,
	})
}

// Dislike 取消点赞接口
func Dislike(ctx *gin.Context) {
	var req LikeOrCollectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := global.Db.First(&user, req.Uid).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  "用户不存在",
		})
		return
	}

	// 检查点赞记录是否存在
	var existingLike models.Like
	if err := global.Db.Where("uid = ? AND nid = ?", req.Uid, req.NoteID).First(&existingLike).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  "未找到点赞记录",
		})
		return
	}

	// 删除 Like 表记录
	if err := global.Db.Delete(&existingLike).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, LikeOrCollectResponse{
			Status: "失败",
			Code:   500,
			Error:  "取消点赞失败",
		})
		return
	}

	// 更新 Note 的 like_count
	if err := global.Db.Model(&models.Note{}).
		Where("note_id = ?", req.NoteID).
		Update("like_counts", gorm.Expr("like_counts - ?", 1)).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, LikeOrCollectResponse{
			Status: "失败",
			Code:   500,
			Error:  "更新 Note 点赞数失败",
		})
		return
	}

	// 更新相关 Tag 的 like_count
	var note models.Note
	if err := global.Db.First(&note, req.NoteID).Error; err == nil {
		tags := strings.Split(note.NoteTagList, ",")
		global.Db.Model(&models.Tag{}).
			Where("t_name IN ?", tags).
			Update("like_count", gorm.Expr("like_count - ?", 1))
	}

	ctx.JSON(http.StatusOK, LikeOrCollectResponse{
		Status: "取消点赞成功",
		Code:   200,
	})
}

// Collect 收藏接口
func Collect(ctx *gin.Context) {
	var req LikeOrCollectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 检查用户是否已经收藏过
	var existingCollect models.Collect
	if err := global.Db.Where("uid = ? AND nid = ?", req.Uid, req.NoteID).First(&existingCollect).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  "您已收藏过此笔记",
		})
		return
	}

	// 添加 Collect 表记录
	collect := models.Collect{
		Uid:        req.Uid,
		Nid:        req.NoteID,
		CreateDate: time.Now(),
	}
	if err := global.Db.Create(&collect).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, LikeOrCollectResponse{
			Status: "失败",
			Code:   500,
			Error:  "收藏失败",
		})
		return
	}

	// 更新 Note 的 collect_count
	if err := global.Db.Model(&models.Note{}).
		Where("note_id = ?", req.NoteID).
		Update("collect_counts", gorm.Expr("collect_counts + ?", 1)).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, LikeOrCollectResponse{
			Status: "失败",
			Code:   500,
			Error:  "更新 Note 收藏数失败",
		})
		return
	}

	// 更新相关 Tag 的 collect_count
	var note models.Note
	if err := global.Db.First(&note, req.NoteID).Error; err == nil {
		tags := strings.Split(note.NoteTagList, ",")
		global.Db.Model(&models.Tag{}).
			Where("t_name IN ?", tags).
			Update("collect_count", gorm.Expr("collect_count + ?", 1))
	}

	// 添加通知记录
	if err := AddNotificationAndUpdateUnreadCount(req.Uid, note.NoteCreatorID, "collect"); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "通知记录创建失败：" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, LikeOrCollectResponse{
		Status: "收藏成功",
		Code:   200,
	})
}

// Uncollect 取消收藏接口
func Uncollect(ctx *gin.Context) {
	var req LikeOrCollectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := global.Db.First(&user, req.Uid).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  "用户不存在",
		})
		return
	}

	// 检查收藏记录是否存在
	var existingCollect models.Collect
	if err := global.Db.Where("uid = ? AND nid = ?", req.Uid, req.NoteID).First(&existingCollect).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, LikeOrCollectResponse{
			Status: "失败",
			Code:   400,
			Error:  "未找到收藏记录",
		})
		return
	}

	// 删除 Collect 表记录
	if err := global.Db.Delete(&existingCollect).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, LikeOrCollectResponse{
			Status: "失败",
			Code:   500,
			Error:  "取消收藏失败",
		})
		return
	}

	// 更新 Note 的 collect_count
	if err := global.Db.Model(&models.Note{}).
		Where("note_id = ?", req.NoteID).
		Update("collect_counts", gorm.Expr("collect_counts - ?", 1)).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, LikeOrCollectResponse{
			Status: "失败",
			Code:   500,
			Error:  "更新 Note 收藏数失败",
		})
		return
	}

	// 更新相关 Tag 的 collect_count
	var note models.Note
	if err := global.Db.First(&note, req.NoteID).Error; err == nil {
		tags := strings.Split(note.NoteTagList, ",")
		global.Db.Model(&models.Tag{}).
			Where("t_name IN ?", tags).
			Update("collect_count", gorm.Expr("collect_count - ?", 1))
	}

	ctx.JSON(http.StatusOK, LikeOrCollectResponse{
		Status: "取消收藏成功",
		Code:   200,
	})
}

// GetIfUserLikeOrCollect 检查用户是否点赞、收藏以及是否关注帖子作者
func GetIfUserLikeOrCollect(ctx *gin.Context) {
	// 获取请求参数
	uid := ctx.Query("uid")
	nid := ctx.Query("nid")

	// 校验参数
	if uid == "" || nid == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少必要参数(uid 或 nid)",
		})
		return
	}

	// 转换参数为整数
	userID, err := strconv.Atoi(uid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "uid 参数格式不正确",
		})
		return
	}

	noteID, err := strconv.Atoi(nid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "nid 参数格式不正确",
		})
		return
	}

	// 调用核心逻辑函数
	likeStatus := utils.CheckIfUserLiked(userID, noteID)
	collectStatus := utils.CheckIfUserCollected(userID, noteID)

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data": gin.H{
			"like":    likeStatus,
			"collect": collectStatus,
		},
	})
}
