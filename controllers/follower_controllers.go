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

// UnfollowRequest 取消关注请求结构
type UnfollowRequest struct {
	CurrentUserID uint `json:"current_user_id" binding:"required"` // 当前用户ID
	TargetUserID  uint `json:"target_user_id" binding:"required"`  // 目标用户ID
}

// FollowResponse 关注操作的响应结构
type FollowResponse struct {
	Code    int    `json:"code"`    // 状态码
	Success bool   `json:"success"` // 操作是否成功
	Msg     string `json:"msg"`     // 消息
	FStatus string `json:"fstatus"` // 关注状态，例如 "follows" 或 "unfollows"
}

// Follow 关注接口
// @Summary 用户关注接口
// @Description 当前用户可以通过此接口关注目标用户
// @Tags 关注相关接口
// @Accept application/json
// @Produce application/json
// @Param data body FollowRequest true "关注请求参数"
// @Success 200 {object} FollowResponse "关注成功响应信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /follow [post]
func Follow(ctx *gin.Context) {
	var req FollowRequest

	// 绑定请求体数据
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 防止用户关注自己
	if req.CurrentUserID == req.TargetUserID {
		ctx.JSON(http.StatusBadRequest, FollowResponse{
			Code:    400,
			Success: false,
			Msg:     "不能关注自己",
			FStatus: "",
		})
		return
	}

	// 检查是否已经关注
	var existingFollower models.Follower
	if err := global.Db.Where("uid = ? AND fid = ?", req.CurrentUserID, req.TargetUserID).First(&existingFollower).Error; err == nil {
		ctx.JSON(http.StatusOK, FollowResponse{
			Code:    200,
			Success: true,
			Msg:     "已关注",
			FStatus: "follows",
		})
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
		ctx.JSON(http.StatusInternalServerError, FollowResponse{
			Code:    500,
			Success: false,
			Msg:     "关注失败",
			FStatus: "",
		})
		return
	}

	// 更新计数
	global.Db.Model(&models.User{}).Where("user_id = ?", req.TargetUserID).Update("fan_count", gorm.Expr("fan_count + ?", 1))
	global.Db.Model(&models.User{}).Where("user_id = ?", req.CurrentUserID).Update("follower_count", gorm.Expr("follower_count + ?", 1))

	ctx.JSON(http.StatusOK, FollowResponse{
		Code:    200,
		Success: true,
		Msg:     "成功",
		FStatus: "follows",
	})
}

// Unfollow 取消关注接口
// @Summary 用户取消关注接口
// @Description 当前用户可以通过此接口取消对目标用户的关注
// @Tags 关注相关接口
// @Accept application/json
// @Produce application/json
// @Param data body UnfollowRequest true "取消关注请求参数"
// @Success 200 {object} FollowResponse "取消关注成功响应信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "未找到关注关系"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /unfollow [post]
func Unfollow(ctx *gin.Context) {
	var req UnfollowRequest

	// 绑定请求体数据
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 检查是否存在关注关系
	var existingFollower models.Follower
	if err := global.Db.Where("uid = ? AND fid = ?", req.CurrentUserID, req.TargetUserID).First(&existingFollower).Error; err != nil {
		ctx.JSON(http.StatusOK, FollowResponse{
			Code:    404,
			Success: false,
			Msg:     "未找到关注关系",
			FStatus: "",
		})
		return
	}

	// 删除关注记录
	if err := global.Db.Delete(&existingFollower).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, FollowResponse{
			Code:    500,
			Success: false,
			Msg:     "取消关注失败",
			FStatus: "",
		})
		return
	}

	// 更新计数
	global.Db.Model(&models.User{}).Where("user_id = ?", req.TargetUserID).Update("fan_count", gorm.Expr("fan_count - ?", 1))
	global.Db.Model(&models.User{}).Where("user_id = ?", req.CurrentUserID).Update("follower_count", gorm.Expr("follower_count - ?", 1))

	ctx.JSON(http.StatusOK, FollowResponse{
		Code:    200,
		Success: true,
		Msg:     "成功",
		FStatus: "unfollows",
	})
}
