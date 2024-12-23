package config

import (
	"log"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() {
	dsn := AppCongfig.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//NamingStrategy: schema.NamingStrategy{
		//	NoLowerCase: true, // 禁用下划线风格转换
		//},
	})

	if err != nil {
		log.Fatalf("Fail to initialize database, got error: %v", err)
	}

	// 先迁移 User 表
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Error migrating User table: %v", err)
	}
	// 再迁移 Follower 表
	err = db.AutoMigrate(&models.Follower{})
	if err != nil {
		log.Fatalf("Error migrating Follower table: %v", err)
	}
	// 再迁移 tag 表
	err = db.AutoMigrate(&models.Tag{})
	if err != nil {
		log.Fatalf("Error migrating Tag table: %v", err)
	}
	// 再迁移 note 表
	err = db.AutoMigrate(&models.Note{})
	if err != nil {
		log.Fatalf("Error migrating Note table: %v", err)
	}
	// 再迁移 tag_note_relation 表
	err = db.AutoMigrate(&models.TagNoteRelation{})
	if err != nil {
		log.Fatalf("Error migrating tag_note_relation table: %v", err)
	}
	// 再迁移 comments 表
	err = db.AutoMigrate(&models.Comments{})
	if err != nil {
		log.Fatalf("Error migrating comments table: %v", err)
	}
	// 再迁移 like 表
	err = db.AutoMigrate(&models.Like{})
	if err != nil {
		log.Fatalf("Error migrating like table: %v", err)
	}
	// 再迁移 collect 表
	err = db.AutoMigrate(&models.Collect{})
	if err != nil {
		log.Fatalf("Error migrating collect table: %v", err)
	}
	// 再迁移 Notification 表
	err = db.AutoMigrate(&models.Notification{})
	if err != nil {
		log.Fatalf("Error migrating Notification table: %v", err)
	}

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
