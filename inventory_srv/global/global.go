package global

import (
	"gorm.io/gorm"

	"github.com/xlt/shop_srv/inventory_srv/config"
)

var (
	ServerConfig = &config.ServerConfig{}
	MySQLConn    *gorm.DB
	NacosConfig  = &config.NacosConfig{}
)
