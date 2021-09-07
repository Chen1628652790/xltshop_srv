package main

import (
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/xlt/shop_srv/user_srv/model"
)

func main() {
	dsn := "root:root@tcp(192.168.0.105:3306)/xltshop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)

	// 全局模式
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("gorm.Open failed. err:", err.Error())
	}

	options := &password.Options{
		SaltLen:      16,
		Iterations:   100,
		KeyLen:       32,
		HashFunction: sha512.New,
	}
	salt, encodePwd := password.Encode("admin123", options)
	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodePwd)

	for i := 0; i < 10; i++ {
		user := model.User{
			Mobile:   fmt.Sprintf("1510000111%d", i),
			Password: newPassword,
			NickName: fmt.Sprintf("xiaolatiao%d", i),
		}
		db.Create(&user)
	}

	//err = db.AutoMigrate(&model.User{})
	//if err != nil {
	//	log.Fatal("db.AutoMigrate failed. err:", err.Error())
	//}
}
