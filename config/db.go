package config

import (
	"log"
	"time"
	"travel-from-sysu-backend/global"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() {
	dsn := AppCongfig.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Fail to initialize database, got error: %v", err)
	}

	sqlDB, err := db.DB()

	sqlDB.SetMaxIdleConns(AppCongfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(AppCongfig.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err != nil {
		log.Fatalf("Fail to configure database, got error: %v", err)
	}

	global.Db = db
}
