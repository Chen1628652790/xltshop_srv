package global

import (
	"github.com/xlt/shop_srv/order_srv/proto"
	"gorm.io/gorm"

	"github.com/xlt/shop_srv/order_srv/config"
)

var (
	ServerConfig = &config.ServerConfig{}
	MySQLConn    *gorm.DB
	NacosConfig  = &config.NacosConfig{}

	GoodsServer     proto.GoodsClient
	InventoryServer proto.InventoryClient
)
