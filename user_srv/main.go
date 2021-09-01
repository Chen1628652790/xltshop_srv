package main

import (
	"flag"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

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

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		zap.S().Errorw("net.Listen failed, err:", "msg", err.Error())
		return
	}

	err = server.Serve(listen)
	if err != nil {
		zap.S().Errorw("server.Serve failed, err:", "msg", err.Error())
	}
}
