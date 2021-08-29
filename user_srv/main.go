package main

import (
	"flag"
	"fmt"
	"github.com/xlt/shop_srv/user_srv/handler"
	"github.com/xlt/shop_srv/user_srv/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int64("port", 50051, "端口号")
	flag.Parse()

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		log.Fatal("net.Listen failed, err:", err.Error())
	}

	err = server.Serve(listen)
	if err != nil {
		log.Fatal("server.Serve failed, err:", err.Error())
	}
}
