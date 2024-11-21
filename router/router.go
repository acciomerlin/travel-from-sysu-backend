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
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}
	user := r.Group("/api/user")
	{
		user.POST("/follow", controllers.Follow)
		user.POST("/unfollow", controllers.Unfollow)
	}
	return r
}
