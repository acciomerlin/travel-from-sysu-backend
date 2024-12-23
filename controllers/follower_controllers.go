package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
	"travel-from-sysu-backend/utils"
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

// UserFoCountsRequest 获取用户关注人数与粉丝数请求结构
type UserFoCountsRequest struct {
	UserID uint `form:"user_id" json:"user_id" binding:"required"`
}

// UserFoCountsResponse 获取用户关注人数与粉丝数响应结构
type UserFoCountsResponse struct {
	Code          int    `json:"code"`           // 状态码
	Success       bool   `json:"success"`        // 是否成功
	Msg           string `json:"msg"`            // 消息
	FollowerCount uint64 `json:"follower_count"` // 关注人数
	FanCount      uint64 `json:"fan_count"`      // 粉丝人数
}

// SimplifiedUser 精简的用户信息返回体
type SimplifiedUser struct {
	UserID        uint   `json:"user_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	FanCount      uint64 `json:"fan_count"`
	FollowerCount uint64 `json:"follower_count"`
	Gender        *int   `json:"gender"`
	Avatar        string `json:"avatar"`
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

	if err := AddNotificationAndUpdateUnreadCount(req.CurrentUserID, req.TargetUserID, "follow"); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "通知记录创建失败：" + err.Error(),
		})
		return
	}

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

// GetUserFoCounts 获取用户关注人数和粉丝人数接口
func GetUserFoCounts(ctx *gin.Context) {
	var req UserFoCountsRequest

	// 绑定请求体数据
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 查询用户信息
	var user models.User
	if err := global.Db.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{
				Status: "失败",
				Code:   404,
				Error:  "未找到用户",
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{
				Status: "失败",
				Code:   500,
				Error:  "服务器内部错误",
			})
		}
		return
	}

	// 返回统计信息
	ctx.JSON(http.StatusOK, UserFoCountsResponse{
		Code:          200,
		Success:       true,
		Msg:           "获取成功",
		FollowerCount: user.FollowerCount,
		FanCount:      user.FanCount,
	})
}

// GetFollowersWithPagination 获取用户关注的用户的部分用户信息，采用游标分页
func GetFollowersWithPagination(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	startCursor := ctx.Query("start_cursor") // 游标
	num := ctx.Query("limit")

	if userID == "" || num == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"msg":     "缺少参数 user_id 或 limit",
		})
		return
	}

	limit := 30 // 默认最多返回30个对象
	if n, err := strconv.Atoi(num); err == nil && n > 0 && n < 30 {
		limit = n
	}

	var followers []models.Follower
	query := global.Db.Where("uid = ?", userID)

	// 使用游标进行分页
	//println("DEBUG: startCursor: ", startCursor)
	if startCursor != "" {
		query = query.Where("id > ?", startCursor) // 游标为上次返回的最后一条 ID
	}

	if err := query.Order("id ASC").Limit(limit).Find(&followers).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询失败",
		})
		return
	}

	// 获取关注人的用户信息
	var userIDs []uint
	for _, follower := range followers {
		//println("DEBUG: folower id: ", follower.Fid)
		userIDs = append(userIDs, follower.Fid)
	}

	var users []models.User
	if len(userIDs) > 0 {
		if err := global.Db.Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"success": false,
				"msg":     "查询用户信息失败",
			})
			return
		}
	}

	// 映射为简化的用户信息
	var simplifiedUsers []SimplifiedUser
	for _, user := range users {
		simplifiedUsers = append(simplifiedUsers, SimplifiedUser{
			UserID:        user.UserId,
			Name:          user.Username,
			Description:   user.Description,
			FanCount:      user.FanCount,
			FollowerCount: user.FollowerCount,
			Gender:        user.Gender,
			Avatar:        user.Avatar,
		})
	}

	// 提取下一个游标
	var nextCursor string
	if len(followers) > 0 {
		nextCursor = strconv.Itoa(int(followers[len(followers)-1].ID))
		//println("GET FOLLOWERS DEBUG: nextCursor = " + nextCursor)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":       200,
		"success":    true,
		"msg":        "获取成功",
		"data":       simplifiedUsers,
		"nextCursor": nextCursor, // 返回游标供前端下次请求使用
	})
}

// GetFolloweesWithPagination 获取用户粉丝的部分用户信息，采用游标分页
func GetFolloweesWithPagination(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	startCursor := ctx.Query("start_cursor") // 游标，用于分页
	num := ctx.Query("limit")                // 每次请求的最大条数

	// 检查参数
	if userID == "" || num == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"msg":     "缺少参数 user_id 或 limit",
		})
		return
	}

	// 设置默认最大限制
	limit := 30
	if n, err := strconv.Atoi(num); err == nil && n > 0 && n < 30 {
		limit = n
	}

	var followees []models.Follower
	query := global.Db.Where("fid = ?", userID)

	// 使用游标进行分页
	if startCursor != "" {
		query = query.Where("id > ?", startCursor) // 游标为上次返回的最后一条记录的 ID
	}

	// 查询数据
	if err := query.Order("id ASC").Limit(limit).Find(&followees).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询失败",
		})
		return
	}

	// 获取粉丝的用户信息
	var userIDs []uint
	for _, followee := range followees {
		userIDs = append(userIDs, followee.Uid)
	}

	var users []models.User
	if len(userIDs) > 0 {
		if err := global.Db.Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"success": false,
				"msg":     "查询用户信息失败",
			})
			return
		}
	}

	// 映射为简化的用户信息
	var simplifiedUsers []SimplifiedUser
	for _, user := range users {
		simplifiedUsers = append(simplifiedUsers, SimplifiedUser{
			UserID:        user.UserId,
			Name:          user.Username,
			Description:   user.Description,
			FanCount:      user.FanCount,
			FollowerCount: user.FollowerCount,
			Gender:        user.Gender,
			Avatar:        user.Avatar,
		})
	}

	// 提取下一个游标
	var nextCursor string
	if len(followees) > 0 {
		nextCursor = strconv.Itoa(int(followees[len(followees)-1].ID)) // 最后一条记录的 ID
		//println("GET FOLLOWEES DEBUG: nextCursor = " + nextCursor)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":       200,
		"success":    true,
		"msg":        "获取成功",
		"data":       simplifiedUsers,
		"nextCursor": nextCursor, // 前端通过 nextCursor 获取下一页数据
	})
}

// GetIfUserFollow 获取用户是否关注帖子作者 （先弃用）
func GetIfUserFollow(ctx *gin.Context) {
	// 获取请求参数
	uid := ctx.Query("uid")
	fid := ctx.Query("fid")

	// 参数校验
	if uid == "" || fid == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少必要参数(用户uid 或 帖子作者fid)",
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

	followID, err := strconv.Atoi(fid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "fid 参数格式不正确",
		})
		return
	}

	// 调用函数获取是否已关注
	followStatus := utils.CheckUserFollow(userID, followID)

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"follow": followStatus,
	})
}
