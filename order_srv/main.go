package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/consul/api"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/xlt/shop_srv/order_srv/global"
	"github.com/xlt/shop_srv/order_srv/initialize"
	"github.com/xlt/shop_srv/order_srv/proto"
	"github.com/xlt/shop_srv/order_srv/utils"
)

func main() {
	//IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int64("port", 0, "端口号")
	flag.Parse()

	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitMySQL()
	initialize.InitClient()

	if *Port == 0 {
		*Port = int64(utils.GetFreePort())
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", global.ServerConfig.Host, *Port))
	if err != nil {
		zap.S().Errorw("net.Listen failed, err:", "msg", err.Error())
		return
	}

	server := grpc.NewServer()
	proto.RegisterOrderServer(server, &proto.UnimplementedOrderServer{})
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// 创建Consul客户端
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port)
	consulClient, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	// 注册order_srv的grpc检查
	check := &api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%s:%d", global.ServerConfig.Host, *Port),
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "10s",
	}

	// uuid 保证服务唯一
	serviceID := fmt.Sprintf("%s", uuid.NewV4())

	// 配置注册信息方便 web 层调用
	registration := new(api.AgentServiceRegistration)
	registration.Name = global.ServerConfig.ServerName
	registration.ID = serviceID
	registration.Port = int(*Port)
	registration.Tags = global.ServerConfig.Tags
	registration.Address = global.ServerConfig.Host
	registration.Check = check
	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		zap.S().Errorw("client.Agent().ServiceRegister failed", "msg", err.Error())
		panic(err)
	}

	zap.S().Infow("server.Serve success", "port", *Port, "serviveID", serviceID)
	go func() {
		err = server.Serve(listen)
		if err != nil {
			zap.S().Errorw("server.Serve failed, err:", "msg", err.Error())
		}
	}()

	// signal
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err := consulClient.Agent().ServiceDeregister(serviceID); err != nil {
		zap.S().Errorw("consulClient.Agent().ServiceDeregister failed", "msg", err.Error())
		return
	}
	zap.S().Infow("注销服务成功", "port", *Port, "serviveID", serviceID)
}
