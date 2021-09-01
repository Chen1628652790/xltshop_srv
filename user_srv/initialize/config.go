package initialize

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/xlt/shop_srv/user_srv/global"
)

func InitConfig() {
	var configFileName string

	if global.ServerConfig.Mode == "debug" {
		configFileName = "config-debug.yaml"
	} else {
		configFileName = "config-pro.yaml"
	}

	v := viper.New()
	v.SetConfigFile(configFileName)
	if err := v.ReadInConfig(); err != nil {
		zap.S().Errorw("v.ReadInConfig failed", "msg", err.Error())
		return
	}

	if err := v.Unmarshal(global.ServerConfig); err != nil {
		zap.S().Errorw("v.Unmarshal failed", "msg", err.Error())
		return
	}
	zap.S().Infow("配置文件读取成功", "data", global.ServerConfig)

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		zap.S().Infow("配置文件发生变化", "filename", e.Name)
		if err := v.ReadInConfig(); err != nil {
			zap.S().Errorw("v.ReadInConfig failed", "msg", err.Error())
			return
		}

		if err := v.Unmarshal(global.ServerConfig); err != nil {
			zap.S().Errorw("v.Unmarshal failed", "msg", err.Error())
			return
		}
		zap.S().Infow("配置文件重新读取成功", "data", global.ServerConfig)
	})
}
