package initialize

import (
	"fmt"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/xlt/shop_srv/order_srv/global"
	"github.com/xlt/shop_srv/order_srv/proto"
)

func InitClient() {
	initInventoryClientLoadBalance()
	initGoodsClientLoadBalance()
}

func initInventoryClientLoadBalance() {
	inventoryConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s",
			global.ServerConfig.ConsulConfig.Host,
			global.ServerConfig.ConsulConfig.Port,
			global.ServerConfig.InventoryServerConfig.Name,
		),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Errorw("grpc.Dial failed", "msg", err.Error())
	}

	global.InventoryServer = proto.NewInventoryClient(inventoryConn)
}

func initGoodsClientLoadBalance() {
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s",
			global.ServerConfig.ConsulConfig.Host,
			global.ServerConfig.ConsulConfig.Port,
			global.ServerConfig.GoodsServerConfig.Name,
		),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Errorw("grpc.Dial failed", "msg", err.Error())
	}

	global.GoodsServer = proto.NewGoodsClient(goodsConn)
}
