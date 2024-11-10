package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string
		Port string
	}
	Database struct {
		Dsn          string
		MaxIdleConns int
		MaxOpenConns int
	}
}

var AppCongfig *Config

func InitConfig() {
	//配置
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")

	//read config
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	AppCongfig = &Config{}

	if err := viper.Unmarshal(AppCongfig); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	InitDB()
}
