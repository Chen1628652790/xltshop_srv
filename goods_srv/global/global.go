package global

import (
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"

	"github.com/xlt/shop_srv/goods_srv/config"
)

var (
	ServerConfig = &config.ServerConfig{}
	MySQLConn    *gorm.DB
	NacosConfig  = &config.NacosConfig{}
	EsClient     = &elastic.Client{}
)
