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

// ReadNotifications 阅读消息接口
func ReadNotifications(ctx *gin.Context) {
	// 获取用户ID
	userID := ctx.Query("user_id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少必要参数 user_id",
		})
		return
	}

	// 转换 userID 为整数
	uid, err := strconv.Atoi(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "user_id 参数格式不正确",
		})
		return
	}

	// 查询未读消息
	var unreadNotifications []models.Notification
	if err := global.Db.Where("recipient_id = ? AND is_read = ?", uid, false).Find(&unreadNotifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询未读消息失败",
		})
		return
	}

	// 查询已读消息
	var readNotifications []models.Notification
	if err := global.Db.Where("recipient_id = ? AND is_read = ?", uid, true).Find(&readNotifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "查询已读消息失败",
		})
		return
	}

	// 将未读消息标记为已读
	if len(unreadNotifications) > 0 {
		if err := global.Db.Model(&models.Notification{}).
			Where("recipient_id = ? AND is_read = ?", uid, false).
			Update("is_read", true).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "更新未读消息为已读失败",
			})
			return
		}
	}

	// 更新用户表中的未读消息数量为 0
	if err := global.Db.Model(&models.User{}).
		Where("user_id = ?", uid).
		Update("unread_noti_count", 0).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "更新用户未读消息数量失败",
		})
		return
	}

	// 构造返回数据
	response := gin.H{
		"unread_messages": unreadNotifications,
		"read_messages":   readNotifications,
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"data":   response,
	})
}
