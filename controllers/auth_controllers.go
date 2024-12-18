package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
	"travel-from-sysu-backend/oss"
	"travel-from-sysu-backend/utils"
)

// UserRegisterRequest 注册请求参数
type UserRegisterRequest struct {
	Username    string `json:"username" example:"user123" binding:"required"`
	Password    string `json:"password" example:"password123" binding:"required"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

// RegisterResponse 注册成功的返回信息
type RegisterResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Token  string      `json:"token"`
	User   models.User `json:"user"`
}

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录成功响应体
type LoginResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Token  string      `json:"token"`
	User   models.User `json:"user"`
}

// ChangePwdRequest 修改密码请求体
type ChangePwdRequest struct {
	Username    string `json:"username" binding:"required"`     // 用户名
	OldPassword string `json:"old_password" binding:"required"` // 旧密码
	NewPassword string `json:"new_password" binding:"required"` // 新密码
}

// ChangePwdResponse 修改密码成功响应体
type ChangePwdResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error"`
}

// ChangeUserInfoRequest 修改用户名请求体
type ChangeUserInfoRequest struct {
	Username    string `json:"username" binding:"required"` // 旧用户名
	NewUsername string `json:"new_username"`                // 新用户名
	Description string `json:"description"`
	Gender      *int   `json:"gender"`
	Birthday    string `json:"birthday"`
}

// ChangeUserInfoResponse 修改用户名响应体
type ChangeUserInfoResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error"`
}

// GetUserInfoByIDResponse 查找用户名成功的返回信息
type GetUserInfoByIDResponse struct {
	Status        string    `json:"status"`
	Code          int       `json:"code"`
	Username      string    `json:"username"`
	CreatedAt     time.Time `json:"created_at"`
	Description   string    `json:"description"`
	Gender        *int      `json:"gender"`
	Birthday      string    `json:"birthday"`
	TrendCount    uint64    `json:"trand_count"`
	FollowerCount uint64    `json:"follower_count"`
	FanCount      uint64    `json:"fan_count"`

	Error string `json:"error,omitempty"`
}

// GetNoteCountsByIDResponse 查找用户发帖数成功的返回信息
type GetNoteCountsByIDResponse struct {
	Status    string `json:"status"`
	Code      int    `json:"code"`
	NoteCount uint64 `json:"note_count"`
	Error     string `json:"error,omitempty"`
}

// Register @ChangePwd 修改密码接口
// @Register 用户注册接口
// @Summary 用户注册接口
// @Description 用户注册，接收用户名和密码并生成用户账号
// @Tags 权限相关接口
// @Accept application/json
// @Produce application/json
// @Param user body UserRegisterRequest true "用户注册信息"
// @Success 200 {object} RegisterResponse "注册成功返回信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /register [post]
func Register(ctx *gin.Context) {
	var req UserRegisterRequest

	// 绑定 JSON 数据到 UserRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 创建完整的 User 模型实例
	user := models.User{
		Username:    req.Username,
		Password:    req.Password,
		Phone:       req.Phone,
		Email:       req.Email,
		Description: req.Description,
	}

	// 生成哈希密码
	hashedPwd, err := utils.HashPwd(user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  err.Error(),
		})
		return
	}
	user.Password = hashedPwd

	// 生成 JWT 令牌
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  err.Error(),
		})
		return
	}

	//// 数据库自动迁移
	//if err := global.Db.AutoMigrate(&user); err != nil {
	//	ctx.JSON(http.StatusInternalServerError, ErrorResponse{
	//		Status: "数据库迁移失败",
	//		Code:   500,
	//		Error:  err.Error(),
	//	})
	//	return
	//}

	// 检查是否存在重复的电话或邮箱
	var existingUser models.User
	if user.Phone != "" {
		// 如果手机号不为空，检查是否已有相同手机号的用户
		if err := global.Db.Where("phone = ?", user.Phone).First(&existingUser).Error; err == nil {
			// 如果查到已有用户，返回手机号已注册错误
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Status: "失败",
				Code:   400,
				Error:  "手机号已被注册",
			})
			return
		}
	}

	if user.Email != "" {
		// 如果邮箱不为空，检查是否已有相同邮箱的用户
		if err := global.Db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			// 如果查到已有用户，返回邮箱已注册错误
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Status: "失败",
				Code:   400,
				Error:  "邮箱已被注册",
			})
			return
		}
	}

	if user.Username != "" {
		// 如果用户名不为空，检查是否已有相同用户名的用户
		if err := global.Db.Where("username = ?", user.Email).First(&existingUser).Error; err == nil {
			// 如果查到已有用户，返回用户名已注册错误
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Status: "失败",
				Code:   400,
				Error:  "用户名已被注册",
			})
			return
		}
	}

	// 将用户记录插入数据库
	if err := global.Db.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  err.Error(),
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, RegisterResponse{
		Status: "注册成功",
		Code:   200,
		Token:  token,
		User:   user,
	})
}

// Login 用户登录接口
func Login(ctx *gin.Context) {
	var req LoginRequest

	// 绑定 JSON 数据到 LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 查找用户
	var user models.User
	if err := global.Db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Status: "失败",
			Code:   401,
			Error:  "该用户不存在",
		})
		return
	}

	// 比较密码
	if err := utils.CheckPwd(user.Password, req.Password); err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Status: "失败",
			Code:   401,
			Error:  "密码错误",
		})
		return
	}

	// 生成 JWT 令牌
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  err.Error(),
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, LoginResponse{
		Status: "登录成功",
		Code:   200,
		Token:  token,
		User:   user,
	})
}

// ChangePwd 修改密码接口
func ChangePwd(ctx *gin.Context) {
	var req ChangePwdRequest

	// 绑定 JSON 数据到 ChangePwdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 查找用户
	var user models.User
	if err := global.Db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Status: "失败",
			Code:   401,
			Error:  "用户不存在",
		})
		return
	}

	// 验证旧密码
	if err := utils.CheckPwd(user.Password, req.OldPassword); err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Status: "失败",
			Code:   401,
			Error:  "旧密码错误",
		})
		return
	}

	// 生成新密码哈希值
	hashedPwd, err := utils.HashPwd(req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "密码加密失败",
		})
		return
	}

	// 更新数据库中的密码
	user.Password = hashedPwd
	if err := global.Db.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "密码更新失败",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, ChangePwdResponse{
		Status: "密码修改成功",
		Code:   200,
	})
}

// ChangeUserInfo 修改用户信息接口
func ChangeUserInfo(ctx *gin.Context) {
	var req ChangeUserInfoRequest

	// 绑定 JSON 数据到 ChangeNameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 查找用户
	var user models.User
	if err := global.Db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Status: "失败",
			Code:   401,
			Error:  "用户不存在",
		})
		return
	}

	user.Username = req.NewUsername
	user.Description = req.Description
	user.Birthday = req.Birthday
	user.Gender = req.Gender
	// 更新数据库中的用户名
	if err := global.Db.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "用户信息修改失败",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, ChangeUserInfoResponse{
		Status: "用户信息修改成功",
		Code:   200,
	})
}

// GetGetUserInfoByID 根据用户ID获取用户信息接口
// @Summary 根据用户ID获取用户名接口
// @Description 根据提供的用户ID查找对应的用户名
// @Tags 用户相关接口
// @Accept application/json
// @Produce application/json
// @Param id query string true "用户ID"
// @Success 200 {object} GetNameByIDResponse "用户名查找成功响应信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "用户未找到"
// @Router /getNameByID [get]
func GetUserInfoByID(ctx *gin.Context) {
	// 从查询字符串中获取参数
	id := ctx.DefaultQuery("id", "")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  "缺少id参数",
		})
		return
	}

	// 查找用户
	var user models.User
	if err := global.Db.Where("user_id = ?", id).First(&user).Error; err != nil {
		// 如果没有找到对应的用户
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Status: "失败",
			Code:   404,
			Error:  "用户未找到",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, GetUserInfoByIDResponse{
		Status:        "成功",
		Code:          200,
		Username:      user.Username,
		CreatedAt:     user.CreatedAt,
		Description:   user.Description,
		Gender:        user.Gender,
		Birthday:      user.Birthday,
		TrendCount:    user.TrendCount,
		FollowerCount: user.FollowerCount,
		FanCount:      user.FanCount,
	})
}

// UploadAvatar 用户头像上传接口
func UploadAvatar(ctx *gin.Context) {
	// 获取用户ID
	uid := ctx.PostForm("uid")
	if uid == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	// 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败: " + err.Error()})
		return
	}

	// 文件大小<5mb
	const minFileSize = 5 * 1024 * 1024 // 5MB
	if file.Size > minFileSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("文件大小不能小于 %dMB", minFileSize/(1024*1024))})
		return
	}

	// 校验文件类型
	allowedExtensions := []string{".jpg", ".jpeg", ".png", ".webp"}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	isAllowed := false
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件类型不支持，仅支持 jpg、png、webp 格式"})
		return
	}

	// 将文件上传到阿里云 OSS
	filePath, err := oss.UploadAliyunOss(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "OSS上传失败: " + err.Error()})
		return
	}

	// 更新用户表的 avatar 字段
	var user models.User
	if err := global.Db.First(&user, uid).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}
	user.Avatar = filePath
	if err := global.Db.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户头像失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "头像上传成功",
		"avatar":  filePath,
	})
}

// GetAvatar 获取用户头像
func GetAvatar(ctx *gin.Context) {
	// 获取用户ID
	uid := ctx.Query("uid")
	if uid == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	// 查找用户
	var user models.User
	if err := global.Db.Where("user_id = ?", uid).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	// 检查是否设置了头像
	if user.Avatar == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "用户未设置头像"})
		return
	}

	// 返回头像URL
	ctx.JSON(http.StatusOK, gin.H{
		"message": "获取头像成功",
		"avatar":  user.Avatar,
	})
}

func GetNoteCountsByID(ctx *gin.Context) {
	// 从查询字符串中获取参数
	id := ctx.DefaultQuery("id", "")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  "缺少id参数",
		})
		return
	}

	// 查找用户
	var user models.User
	if err := global.Db.Where("user_id = ?", id).First(&user).Error; err != nil {
		// 如果没有找到对应的用户
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Status: "失败",
			Code:   404,
			Error:  "用户未找到",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, GetNoteCountsByIDResponse{
		Status:    "成功",
		Code:      200,
		NoteCount: user.NoteCount,
	})
}
