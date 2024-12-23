package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
)

// AddNotificationAndUpdateUnreadCount 添加通知记录并增加未读消息计数
func AddNotificationAndUpdateUnreadCount(initiatorID uint, recipientID uint, notifType string) error {
	// 创建通知记录
	notification := models.Notification{
		InitiatorID: initiatorID,
		RecipientID: recipientID,
		Type:        notifType,
		InitiatedAt: time.Now(),
		IsRead:      false,
	}

	// 插入通知记录
	if err := global.Db.Create(&notification).Error; err != nil {
		return err
	}

	// 更新未读消息计数
	if err := global.Db.Model(&models.User{}).
		Where("user_id = ?", recipientID).
		Update("unread_noti_count", gorm.Expr("unread_noti_count + ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

// GetUnreadNotificationCount 获取用户未读消息数量
func GetUnreadNotificationCount(ctx *gin.Context) {
	uid := ctx.Query("uid")
	if uid == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少用户 ID",
		})
		return
	}

	userID, err := strconv.Atoi(uid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "用户 ID 格式错误",
		})
		return
	}

	// 查询用户的未读消息数量
	var user models.User
	if err := global.Db.Select("unread_noti_count").Where("user_id = ?", userID).First(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询未读消息数量失败：" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data": gin.H{
			"unread_noti_count": user.UnreadNotiCount,
		},
	})
}

func GetUnreadCommentNotifications(ctx *gin.Context) {
	recipientID := ctx.Query("recipient_id") // 用户ID
	cursor := ctx.Query("cursor")            // 游标
	num := ctx.DefaultQuery("num", "10")     // 分页数量，默认10

	// 参数校验
	if recipientID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少 recipient_id 参数",
		})
		return
	}

	recipientIDUint, err := strconv.ParseUint(recipientID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "recipient_id 参数格式错误",
		})
		return
	}

	limit, err := strconv.Atoi(num)
	if err != nil || limit <= 0 {
		limit = 10 // 默认返回10条
	}

	// 构造查询
	query := global.Db.Table("notifications").Where("recipient_id = ? AND type = ? AND is_read = ?", recipientIDUint, "comment", false)

	// 如果提供了游标，则查询小于游标的消息
	if cursor != "" {
		cursorID, err := strconv.Atoi(cursor)
		if err == nil {
			query = query.Where("id < ?", cursorID)
		}
	}

	var notifications []models.Notification
	if err := query.Order("id DESC").Limit(limit).Find(&notifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询未读评论消息失败：" + err.Error(),
		})
		return
	}

	// 标记消息为已读
	if len(notifications) > 0 {
		ids := make([]uint, len(notifications))
		for i, notification := range notifications {
			ids[i] = notification.ID
		}
		// 更新用户表中的未读消息计数
		if err := global.Db.Model(&models.User{}).
			Where("user_id = ?", recipientIDUint).
			Update("unread_noti_count", gorm.Expr("unread_noti_count - ?", len(ids))).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "更新用户未读消息计数失败：" + err.Error(),
			})
			return
		}
		// 批量更新
		if err := global.Db.Model(&models.Notification{}).Where("id IN ?", ids).Update("is_read", true).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "更新消息状态失败：" + err.Error(),
			})
			return
		}
	}

	// 获取下一页游标
	nextCursor := ""
	if len(notifications) > 0 {
		nextCursor = strconv.Itoa(int(notifications[len(notifications)-1].ID))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data": gin.H{
			"notifications": notifications,
			"next_cursor":   nextCursor,
		},
	})
}

func GetUnreadLikeAndCollectNotifications(ctx *gin.Context) {
	recipientID := ctx.Query("recipient_id") // 用户ID
	cursor := ctx.Query("cursor")            // 游标
	num := ctx.DefaultQuery("num", "10")     // 分页数量，默认10

	// 参数校验
	if recipientID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少 recipient_id 参数",
		})
		return
	}

	recipientIDUint, err := strconv.ParseUint(recipientID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "recipient_id 参数格式错误",
		})
		return
	}

	limit, err := strconv.Atoi(num)
	if err != nil || limit <= 0 {
		limit = 10 // 默认返回10条
	}

	// 构造查询
	query := global.Db.Table("notifications").Where("recipient_id = ? AND type IN (?, ?) AND is_read = ?", recipientIDUint, "like", "collect", false)

	// 如果提供了游标，则查询小于游标的消息
	if cursor != "" {
		cursorID, err := strconv.Atoi(cursor)
		if err == nil {
			query = query.Where("id < ?", cursorID)
		}
	}

	var notifications []models.Notification
	if err := query.Order("id DESC").Limit(limit).Find(&notifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询未读点赞+收藏消息失败：" + err.Error(),
		})
		return
	}

	// 标记消息为已读
	if len(notifications) > 0 {
		ids := make([]uint, len(notifications))
		for i, notification := range notifications {
			ids[i] = notification.ID
		}
		// 更新用户表中的未读消息计数
		if err := global.Db.Model(&models.User{}).
			Where("user_id = ?", recipientIDUint).
			Update("unread_noti_count", gorm.Expr("unread_noti_count - ?", len(ids))).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "更新用户未读消息计数失败：" + err.Error(),
			})
			return
		}
		// 批量更新
		if err := global.Db.Model(&models.Notification{}).Where("id IN ?", ids).Update("is_read", true).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "更新消息状态失败：" + err.Error(),
			})
			return
		}
	}

	// 获取下一页游标
	nextCursor := ""
	if len(notifications) > 0 {
		nextCursor = strconv.Itoa(int(notifications[len(notifications)-1].ID))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data": gin.H{
			"notifications": notifications,
			"next_cursor":   nextCursor,
		},
	})
}

func GetNewFollowNotifications(ctx *gin.Context) {
	recipientID := ctx.Query("recipient_id") // 用户ID
	cursor := ctx.Query("cursor")            // 游标
	num := ctx.DefaultQuery("num", "10")     // 分页数量，默认10

	// 参数校验
	if recipientID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少 recipient_id 参数",
		})
		return
	}

	recipientIDUint, err := strconv.ParseUint(recipientID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "recipient_id 参数格式错误",
		})
		return
	}

	limit, err := strconv.Atoi(num)
	if err != nil || limit <= 0 {
		limit = 10 // 默认返回10条
	}

	// 构造查询
	query := global.Db.Table("notifications").Where("recipient_id = ? AND type = ? AND is_read = ?", recipientIDUint, "follow", false)

	// 如果提供了游标，则查询小于游标的消息
	if cursor != "" {
		cursorID, err := strconv.Atoi(cursor)
		if err == nil {
			query = query.Where("id < ?", cursorID)
		}
	}

	var notifications []models.Notification
	if err := query.Order("id DESC").Limit(limit).Find(&notifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询新增关注消息失败：" + err.Error(),
		})
		return
	}

	// 标记消息为已读
	if len(notifications) > 0 {
		ids := make([]uint, len(notifications))
		for i, notification := range notifications {
			ids[i] = notification.ID
		}
		// 更新用户表中的未读消息计数
		if err := global.Db.Model(&models.User{}).
			Where("user_id = ?", recipientIDUint).
			Update("unread_noti_count", gorm.Expr("unread_noti_count - ?", len(ids))).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "更新用户未读消息计数失败：" + err.Error(),
			})
			return
		}
		// 批量更新
		if err := global.Db.Model(&models.Notification{}).Where("id IN ?", ids).Update("is_read", true).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "更新消息状态失败：" + err.Error(),
			})
			return
		}
	}

	// 获取下一页游标
	nextCursor := ""
	if len(notifications) > 0 {
		nextCursor = strconv.Itoa(int(notifications[len(notifications)-1].ID))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data": gin.H{
			"notifications": notifications,
			"next_cursor":   nextCursor,
		},
	})
}

func GetReadCommentNotifications(ctx *gin.Context) {
	recipientID := ctx.Query("recipient_id") // 用户ID
	cursor := ctx.Query("cursor")            // 游标
	num := ctx.DefaultQuery("num", "10")     // 每页数量，默认10

	// 参数校验
	if recipientID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少 recipient_id 参数",
		})
		return
	}

	recipientIDUint, err := strconv.ParseUint(recipientID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "recipient_id 参数格式错误",
		})
		return
	}

	limit, err := strconv.Atoi(num)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// 查询条件：获取已读的评论消息
	query := global.Db.Table("notifications").Where("recipient_id = ? AND type = ? AND is_read = ?", recipientIDUint, "comment", true)

	if cursor != "" {
		cursorID, err := strconv.Atoi(cursor)
		if err == nil {
			query = query.Where("id < ?", cursorID)
		}
	}

	var notifications []models.Notification
	if err := query.Order("id DESC").Limit(limit).Find(&notifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询已读评论消息失败：" + err.Error(),
		})
		return
	}

	// 获取下一页游标
	nextCursor := ""
	if len(notifications) > 0 {
		nextCursor = strconv.Itoa(int(notifications[len(notifications)-1].ID))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data": gin.H{
			"notifications": notifications,
			"next_cursor":   nextCursor,
		},
	})
}

func GetReadLikeAndCollectNotifications(ctx *gin.Context) {
	recipientID := ctx.Query("recipient_id")
	cursor := ctx.Query("cursor")
	num := ctx.DefaultQuery("num", "10")

	// 参数校验
	if recipientID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少 recipient_id 参数",
		})
		return
	}

	recipientIDUint, err := strconv.ParseUint(recipientID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "recipient_id 参数格式错误",
		})
		return
	}

	limit, err := strconv.Atoi(num)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// 查询条件：获取已读的点赞和收藏消息
	query := global.Db.Table("notifications").Where("recipient_id = ? AND type IN (?, ?) AND is_read = ?", recipientIDUint, "like", "collect", true)

	if cursor != "" {
		cursorID, err := strconv.Atoi(cursor)
		if err == nil {
			query = query.Where("id < ?", cursorID)
		}
	}

	var notifications []models.Notification
	if err := query.Order("id DESC").Limit(limit).Find(&notifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询已读点赞和收藏消息失败：" + err.Error(),
		})
		return
	}

	// 获取下一页游标
	nextCursor := ""
	if len(notifications) > 0 {
		nextCursor = strconv.Itoa(int(notifications[len(notifications)-1].ID))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data": gin.H{
			"notifications": notifications,
			"next_cursor":   nextCursor,
		},
	})
}

func GetReadFollowNotifications(ctx *gin.Context) {
	recipientID := ctx.Query("recipient_id")
	cursor := ctx.Query("cursor")
	num := ctx.DefaultQuery("num", "10")

	// 参数校验
	if recipientID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少 recipient_id 参数",
		})
		return
	}

	recipientIDUint, err := strconv.ParseUint(recipientID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "recipient_id 参数格式错误",
		})
		return
	}

	limit, err := strconv.Atoi(num)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// 查询条件：获取已读的关注消息
	query := global.Db.Table("notifications").Where("recipient_id = ? AND type = ? AND is_read = ?", recipientIDUint, "follow", true)

	if cursor != "" {
		cursorID, err := strconv.Atoi(cursor)
		if err == nil {
			query = query.Where("id < ?", cursorID)
		}
	}

	var notifications []models.Notification
	if err := query.Order("id DESC").Limit(limit).Find(&notifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询已读关注消息失败：" + err.Error(),
		})
		return
	}

	// 获取下一页游标
	nextCursor := ""
	if len(notifications) > 0 {
		nextCursor = strconv.Itoa(int(notifications[len(notifications)-1].ID))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data": gin.H{
			"notifications": notifications,
			"next_cursor":   nextCursor,
		},
	})
}
