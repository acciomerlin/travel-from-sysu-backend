package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
	"travel-from-sysu-backend/utils"
)

// RegisterResponse 注册成功的返回信息
type RegisterResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Token  string      `json:"token"`
	User   models.User `json:"user"`
}

// ErrorResponse 错误返回信息
type ErrorResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error"`
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

// Register 用户注册接口
// @Summary 用户注册接口
// @Description 用户注册，接收用户名和密码并生成用户账号
// @Tags 权限相关接口
// @Accept application/json
// @Produce application/json
// @Param user body models.UserRegisterRequest true "用户注册信息"
// @Success 200 {object} RegisterResponse "注册成功返回信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /register [post]
func Register(ctx *gin.Context) {
	var req models.UserRegisterRequest

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

	// 数据库自动迁移
	if err := global.Db.AutoMigrate(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  err.Error(),
		})
		return
	}

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

	// Login 用户登录接口
}

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
