package handler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/xlt/shop_srv/order_srv/global"
	"github.com/xlt/shop_srv/order_srv/model"
	"github.com/xlt/shop_srv/order_srv/proto"
)

type OrderServer struct {
	proto.UnimplementedOrderServer
}

func (*OrderServer) CartItemList(ctx context.Context, req *proto.UserInfo) (*proto.CartItemListResponse, error) {
	//获取用户的购物车列表
	var shopCarts []model.ShoppingCart
	var rsp proto.CartItemListResponse

	if result := global.MySQLConn.Where(&model.ShoppingCart{User: req.Id}).Find(&shopCarts); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, result.Error
	} else {
		rsp.Total = int32(result.RowsAffected)
	}

	for _, shopCart := range shopCarts {
		rsp.Data = append(rsp.Data, &proto.ShopCartInfoResponse{
			Id:      shopCart.ID,
			UserId:  shopCart.User,
			GoodsId: shopCart.Goods,
			Nums:    shopCart.Nums,
			Checked: shopCart.Checked,
		})
	}
	return &rsp, nil
}

func (*OrderServer) CreateCartItem(ctx context.Context, req *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error) {
	//将商品添加到购物车 1. 购物车中原本没有这件商品 - 新建一个记录 2. 这个商品之前添加到了购物车- 合并
	var shopCart model.ShoppingCart

	if result := global.MySQLConn.Where(&model.ShoppingCart{Goods: req.GoodsId, User: req.UserId}).First(&shopCart); result.RowsAffected == 1 {
		//如果记录已经存在，则合并购物车记录, 更新操作
		shopCart.Nums += req.Nums
	} else {
		//插入操作
		shopCart.User = req.UserId
		shopCart.Goods = req.GoodsId
		shopCart.Nums = req.Nums
		shopCart.Checked = false
	}

	if result := global.MySQLConn.Save(&shopCart); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
		return nil, result.Error
	}
	return &proto.ShopCartInfoResponse{Id: shopCart.ID}, nil
}

func (*OrderServer) UpdateCartItem(ctx context.Context, req *proto.CartItemRequest) (*empty.Empty, error) {
	//更新购物车记录，更新数量和选中状态
	var shopCart model.ShoppingCart

	if result := global.MySQLConn.Where("goods=? and user=?", req.GoodsId, req.UserId).First(&shopCart); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}

	shopCart.Checked = req.Checked
	if req.Nums > 0 {
		shopCart.Nums = req.Nums
	}
	if result := global.MySQLConn.Save(&shopCart); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
		return nil, result.Error
	}

	return &empty.Empty{}, nil
}

func (*OrderServer) DeleteCartItem(ctx context.Context, req *proto.CartItemRequest) (*empty.Empty, error) {
	if result := global.MySQLConn.Where("goods=? and user=?", req.GoodsId, req.UserId).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}
	return &empty.Empty{}, nil
}

func (*OrderServer) OrderList(ctx context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orders []model.OrderInfo
	var rsp proto.OrderListResponse

	var total int64
	if result := global.MySQLConn.Model(&model.OrderInfo{}).Where(&model.OrderInfo{User: req.UserId}).Count(&total); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "查询订单列表总数失败")
	}
	rsp.Total = int32(total)

	//分页
	if result := global.MySQLConn.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Where(&model.OrderInfo{User: req.UserId}).Find(&orders); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Scopes failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "查询订单列表失败")
	}
	for _, order := range orders {
		rsp.Data = append(rsp.Data, &proto.OrderInfoResponse{
			Id:      order.ID,
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Total:   order.OrderMount,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SingerMobile,
			AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &rsp, nil
}

func (*OrderServer) OrderDetail(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	var order model.OrderInfo
	var rsp proto.OrderInfoDetailResponse

	//这个订单的id是否是当前用户的订单， 如果在web层用户传递过来一个id的订单， web层应该先查询一下订单id是否是当前用户的
	//在个人中心可以这样做，但是如果是后台管理系统，web层如果是后台管理系统 那么只传递order的id，如果是电商系统还需要一个用户的id
	if result := global.MySQLConn.Where(&model.OrderInfo{BaseModel: model.BaseModel{ID: req.Id}, User: req.UserId}).First(&order); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}

	orderInfo := proto.OrderInfoResponse{}
	orderInfo.Id = order.ID
	orderInfo.UserId = order.User
	orderInfo.OrderSn = order.OrderSn
	orderInfo.PayType = order.PayType
	orderInfo.Status = order.Status
	orderInfo.Post = order.Post
	orderInfo.Total = order.OrderMount
	orderInfo.Address = order.Address
	orderInfo.Name = order.SignerName
	orderInfo.Mobile = order.SingerMobile

	rsp.OrderInfo = &orderInfo

	var orderGoods []model.OrderGoods
	if result := global.MySQLConn.Where(&model.OrderGoods{Order: order.ID}).Find(&orderGoods); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, result.Error
	}

	for _, orderGood := range orderGoods {
		rsp.Goods = append(rsp.Goods, &proto.OrderItemResponse{
			GoodsId:    orderGood.Goods,
			GoodsName:  orderGood.GoodsName,
			GoodsPrice: orderGood.GoodsPrice,
			GoodsImage: orderGood.GoodsImage,
			Nums:       orderGood.Nums,
		})
	}

	return &rsp, nil
}

func (*OrderServer) CreateOrder(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	// 1. 从购物车中查询到选中的商品
	// 2. 查询商品的价格 --- 需要跨服务
	// 3. 库存的扣减 --- 需要阔服务
	// 4. 订单商品信息表
	// 5. 订单基本信息表
	// 6. 从购物车中删除已结算的商品

	var shopCarts []model.ShoppingCart
	if result := global.MySQLConn.Where(&model.ShoppingCart{User: req.UserId, Checked: true}).Find(&shopCarts); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "没有选中结算商品")
	}

	var goodsIDs []int32
	goodsNum := make(map[int32]int32)
	for _, goods := range shopCarts {
		goodsIDs = append(goodsIDs, goods.ID)
		goodsNum[goods.ID] = goods.Nums
	}

	// 商品微服务
	goodsList, err := global.GoodsServer.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{Id: goodsIDs})
	if err != nil {
		zap.S().Errorw("global.GoodsServer.BatchGetGoods failed", "msg", err.Error())
		return nil, err
	}

	var orderAmount float32
	var orderGoods []*model.OrderGoods
	var goodsInvInfo []*proto.GoodsInvInfo
	for _, goods := range goodsList.Data {
		orderAmount += goods.ShopPrice * float32(goodsNum[goods.Id])
		orderGoods = append(orderGoods, &model.OrderGoods{
			Goods:      goods.Id,
			GoodsName:  goods.Name,
			GoodsImage: goods.GoodsFrontImage,
			GoodsPrice: goods.ShopPrice,
			Nums:       goodsNum[goods.Id],
		})
		goodsInvInfo = append(goodsInvInfo, &proto.GoodsInvInfo{
			GoodsId: goods.Id,
			Num:     goodsNum[goods.Id],
		})
	}

	// 库存微服务
	if _, err := global.InventoryServer.Sell(context.Background(), &proto.SellInfo{GoodsInfo: goodsInvInfo}); err != nil {
		zap.S().Errorw("global.InventoryServer.Sell failed", "msg", err.Error())
		return nil, err
	}

	tx := global.MySQLConn.Begin()
	order := model.OrderInfo{
		User:         req.UserId,
		OrderSn:      generateOrderSn(req.UserId),
		OrderMount:   orderAmount,
		Address:      req.Address,
		SignerName:   req.Name,
		SingerMobile: req.Mobile,
		Post:         req.Post,
	}
	if result := tx.Create(&order); result.Error != nil {
		tx.Rollback()
		zap.S().Errorw("global.MySQLConn.Create failed", "msg", err.Error())
		return nil, status.Errorf(codes.Internal, "创建订单失败")
	}

	for _, goods := range orderGoods {
		goods.Order = order.ID
	}
	if result := global.MySQLConn.CreateInBatches(&orderGoods, 20); result.Error != nil {
		tx.Rollback()
		zap.S().Errorw("global.MySQLConn.CreateInBatches failed", "msg", err.Error())
		return nil, status.Errorf(codes.Internal, "创建订单商品失败")
	}

	if result := tx.Where("user = ? and checked = ?", req.UserId, 1).Delete(&model.ShoppingCart{}); result.Error != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "删除购物车商品失败")
	}

	tx.Commit()

	// todo 分布式事务
	return &proto.OrderInfoResponse{
		Id:      order.ID,
		OrderSn: order.OrderSn,
		Total:   order.OrderMount,
	}, nil
}

func (*OrderServer) UpdateOrderStatus(ctx context.Context, req *proto.OrderStatus) (*empty.Empty, error) {
	if result := global.MySQLConn.Model(&model.OrderInfo{}).Select("status").Where("order_sn = ?", req.OrderSn).Update("status = ?", req.Status); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Model(&model.OrderInfo{}).Select failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "更新订单状态失败")
	}

	return &empty.Empty{}, nil
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func generateOrderSn(userID int32) string {
	// 年 +月 + 日 + 时 + 分 + 秒 + 用户ID + 两位随机数
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d%d%d%d%d%d%d%d",
		time.Now().Year(),
		time.Now().Month(),
		time.Now().Day(),
		time.Now().Hour(),
		time.Now().Minute(),
		time.Now().Nanosecond(),
		userID,
		rand.Intn(90)+10)
}
