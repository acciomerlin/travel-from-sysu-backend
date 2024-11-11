package controllers

//权限管理的文件：登录、注册......

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
	"travel-from-sysu-backend/utils"
)

func Register(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error: ": err.Error()})
		return
	}

	hashedPwd, err := utils.HashPwd(user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error: ": err.Error()})
		return
	}

	user.Password = hashedPwd

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error: ": err.Error()})
		return
	}

	if err := global.Db.AutoMigrate(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error: ": err.Error()})
		return
	}

	if err := global.Db.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error: ": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
