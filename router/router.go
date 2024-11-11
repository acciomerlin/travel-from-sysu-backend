package router

//路由管理文件，处理方法移步controllers

import (
	"travel-from-sysu-backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	auth := r.Group("/api/auth")
	{
		//auth.POST("/login", func(ctx *gin.Context) {
		//	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
		//		"msg": "login success",
		//	})
		//})
		//auth.POST("/register", func(ctx *gin.Context) {
		//	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
		//		"msg": "register success",
		//	})
		//})
		auth.POST("/register", controllers.Register)
	}

	return r
}
