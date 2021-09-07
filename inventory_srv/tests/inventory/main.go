package main

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"

	"github.com/xlt/shop_srv/inventory_srv/proto"
)

var invClient proto.InventoryClient
var conn *grpc.ClientConn

func TestSetInv(goodsId, Num int32) {
	_, err := invClient.SetInv(context.Background(), &proto.GoodsInvInfo{
		GoodsId: goodsId,
		Num:     Num,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("设置库存成功")
}

func TestInvDetail(goodsId int32) {
	rsp, err := invClient.InvDetail(context.Background(), &proto.GoodsInvInfo{
		GoodsId: goodsId,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Num)
}

func TestSell(wg *sync.WaitGroup) {
	/*
		1. 第一件扣减成功： 第二件： 1. 没有库存信息 2. 库存不足
		2. 两件都扣减成功
	*/
	defer wg.Done()
	_, err := invClient.Sell(context.Background(), &proto.SellInfo{
		GoodsInfo: []*proto.GoodsInvInfo{
			{GoodsId: 421, Num: 1},
			//{GoodsId: 422, Num: 30},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("库存扣减成功")
}

func TestReback() {
	_, err := invClient.Reback(context.Background(), &proto.SellInfo{
		GoodsInfo: []*proto.GoodsInvInfo{
			{GoodsId: 421, Num: 10},
			//{GoodsId: 422, Num: 30},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("归还成功")
}

func Init() {
	var err error
	conn, err = grpc.Dial("192.168.0.106:63511", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	invClient = proto.NewInventoryClient(conn)
}

func main() {
	Init()

	var wg sync.WaitGroup
	wg.Add(30)
	for i := 0; i < 30; i++ {
		go TestSell(&wg)
	}
	wg.Wait()

	//TestInvDetail(421)
	//TestSell()
	//TestReback()
	conn.Close()
}

//func main() {
//	Init()
//
//	var wg sync.WaitGroup
//	wg.Add(30)
//	for i := 0; i < 30; i++ {
//		todo 并发问题导致库存扣减不正确，因为会出现多个携程同时从数据库中查询库存信息，这个时候拿到的并不是最新的数据
//		go TestSell(&wg)
//	}
//	wg.Wait()
//
//	//TestInvDetail(421)
//	//TestSell()
//	//TestReback()
//	conn.Close()
//}
