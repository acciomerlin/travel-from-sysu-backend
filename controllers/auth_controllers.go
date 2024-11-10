package controllers

import (
	"net/http"
	"travel-from-sysu-backend/models"
	"travel-from-sysu-backend/utils"

	"github.com/gin-gonic/gin"
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
}
