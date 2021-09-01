package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/xlt/shop_srv/user_srv/global"
	"github.com/xlt/shop_srv/user_srv/handler"
	"github.com/xlt/shop_srv/user_srv/initialize"
	"github.com/xlt/shop_srv/user_srv/proto"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int64("port", 50051, "端口号")
	flag.Parse()

	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitMySQL()

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		zap.S().Errorw("net.Listen failed, err:", "msg", err.Error())
		return
	}

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// 创建Consul客户端
	cfg := api.DefaultConfig()
	cfg.Address = "192.168.199.243:8500"
	consulClient, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	// 注册user_srv的grpc检查
	check := &api.AgentServiceCheck{
		GRPC:                           "192.168.199.194:50051",
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "10s",
	}

	// 配置注册信息方便 web 层调用
	registration := new(api.AgentServiceRegistration)
	registration.Name = global.ServerConfig.ServerName
	registration.ID = global.ServerConfig.ServerName
	registration.Port = 50051
	registration.Tags = []string{"xlt", "user", "srv"}
	registration.Address = "192.168.199.194"
	registration.Check = check
	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		zap.S().Errorw("client.Agent().ServiceRegister failed", "msg", err.Error())
		panic(err)
	}

	err = server.Serve(listen)
	if err != nil {
		zap.S().Errorw("server.Serve failed, err:", "msg", err.Error())
	}
}
