package global

import (
	"github.com/xlt/shop_srv/user_srv/config"
	"gorm.io/gorm"
)

var (
	ServerConfig = &config.ServerConfig{}
	DB           *gorm.DB
)
