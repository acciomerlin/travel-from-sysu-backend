package main

//入口文件

import (
	"travel-from-sysu-backend/config"
	"travel-from-sysu-backend/router"
)

func main() {
	config.InitConfig()
	r := router.SetupRouter()

	port := config.AppCongfig.App.Port

	if port == "" {
		port = "9999"
	}

	r.Run(port)
}
