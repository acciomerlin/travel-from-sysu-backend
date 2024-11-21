package controllers

import (
	"gorm.io/gorm"
	"net/http"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"

	"github.com/gin-gonic/gin"
)

// FollowRequest 关注请求结构
type FollowRequest struct {
	CurrentUserID uint `json:"current_user_id" binding:"required"` // 当前用户ID
	TargetUserID  uint `json:"target_user_id" binding:"required"`  // 目标用户ID
}

// Follow 关注接口
func Follow(ctx *gin.Context) {
	var req FollowRequest

	// 绑定请求体数据
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "success": false, "msg": "请求参数错误", "error": err.Error()})
		return
	}

	// 防止用户关注自己
	if req.CurrentUserID == req.TargetUserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "success": false, "msg": "不能关注自己"})
		return
	}

	// 检查是否已经关注
	var existingFollower models.Follower
	if err := global.Db.Where("uid = ? AND fid = ?", req.CurrentUserID, req.TargetUserID).First(&existingFollower).Error; err == nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 200, "success": true, "msg": "已关注", "data": gin.H{"fstatus": "follows"}})
		return
	}

	// 创建关注记录
	follower := models.Follower{
		Uid:       req.CurrentUserID,
		Fid:       req.TargetUserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := global.Db.Create(&follower).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "success": false, "msg": "关注失败"})
		return
	}

	// 更新计数
	global.Db.Model(&models.User{}).Where("user_id = ?", req.TargetUserID).Update("fan_count", gorm.Expr("fan_count + ?", 1))
	global.Db.Model(&models.User{}).Where("user_id = ?", req.CurrentUserID).Update("follower_count", gorm.Expr("follower_count + ?", 1))

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "success": true, "msg": "成功", "data": gin.H{"fstatus": "follows"}})
}

// UnfollowRequest 取消关注请求结构
type UnfollowRequest struct {
	CurrentUserID uint `json:"current_user_id" binding:"required"` // 当前用户ID
	TargetUserID  uint `json:"target_user_id" binding:"required"`  // 目标用户ID
}

// Unfollow 取消关注接口
func Unfollow(ctx *gin.Context) {
	var req UnfollowRequest

	// 绑定请求体数据
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "success": false, "msg": "请求参数错误", "error": err.Error()})
		return
	}

	// 检查是否存在关注关系
	var existingFollower models.Follower
	if err := global.Db.Where("uid = ? AND fid = ?", req.CurrentUserID, req.TargetUserID).First(&existingFollower).Error; err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 404, "success": false, "msg": "未找到关注关系"})
		return
	}

	// 删除关注记录
	if err := global.Db.Delete(&existingFollower).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "success": false, "msg": "取消关注失败"})
		return
	}

	// 更新计数
	global.Db.Model(&models.User{}).Where("user_id = ?", req.TargetUserID).Update("fan_count", gorm.Expr("fan_count - ?", 1))
	global.Db.Model(&models.User{}).Where("user_id = ?", req.CurrentUserID).Update("follower_count", gorm.Expr("follower_count - ?", 1))

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "success": true, "msg": "成功", "data": gin.H{"fstatus": "unfollows"}})
}
