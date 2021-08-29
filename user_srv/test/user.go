package main

import (
	"context"
	"fmt"
	"github.com/xlt/shop_srv/user_srv/proto"
	"google.golang.org/grpc"
	"log"
)

var (
	userClient proto.UserClient
	conn       *grpc.ClientConn
	err        error
)

func main() {
	Init()
	defer conn.Close()

	TestGetUserList()
}

func TestGetUserList() {
	rsp, err := userClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    1,
		PSize: 2,
	})
	if err != nil {
		log.Fatal("userClient.GetUserList failed, err:", err.Error())
	}

	for _, user := range rsp.Data {
		fmt.Println(user.Mobile, user.NickName)
		checkRsp, err := userClient.CheckPassWord(context.Background(), &proto.PassWordCheckInfo{
			Password:          "admin123",
			EncryptedPassword: user.PassWord,
		})
		if err != nil {
			log.Fatal("userClient.CheckPassWord failed, err:", err.Error())
		}
		fmt.Println(checkRsp.Success)
	}
}

func Init() {
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal("grpc.Dial failed, err:", err.Error())
	}

	userClient = proto.NewUserClient(conn)
}
