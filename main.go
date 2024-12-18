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
	"github.com/gin-contrib/cors"
	"time"
	"travel-from-sysu-backend/config"
	"travel-from-sysu-backend/router"
	"travel-from-sysu-backend/utils"
)

func main() {
	go utils.UpdateHotRecommendations()

	config.InitConfig()
	r := router.SetupRouter()

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},                   // 允许的前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},            // 允许的 HTTP 方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // 允许的自定义头
		ExposeHeaders:    []string{"Content-Length"},                          // 可被浏览器访问的头
		AllowCredentials: true,                                                // 是否允许携带 Cookie
		MaxAge:           12 * time.Hour,                                      // 预检请求的缓存时间
	}))

	port := config.AppCongfig.App.Port

	if port == "" {
		port = "9999"
	}

	r.Run(port)
}
