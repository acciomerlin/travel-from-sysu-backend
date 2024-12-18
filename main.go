package main

// @title travel-from-sysu API
// @version 1.0
// @description travel-from-sysu API
// @termsOfService http://swagger.io/terms/

// @contact.name 804
// @contact.url http://www.swagger.io/support  // 可改为公司或项目支持页面的链接
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/auth

import (
	"github.com/gin-gonic/gin"
	"travel-from-sysu-backend/config"
	"travel-from-sysu-backend/router"
)

func main() {
	config.InitConfig()
	r := router.SetupRouter()

	// 配置 CORS
	r.Use(CORSMiddleware())

	port := config.AppCongfig.App.Port

	if port == "" {
		port = "9999"
	}

	r.Run(port)
}

// CORSMiddleware 中间件处理跨域问题
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
