package handler

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xlt/shop_srv/inventory_srv/global"
	"github.com/xlt/shop_srv/inventory_srv/model"
	"github.com/xlt/shop_srv/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

func (hadnler *InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*empty.Empty, error) {
	var inventory model.Inventory
	global.MySQLConn.Where(&model.Inventory{Goods: req.GoodsId}).First(&inventory)
	inventory.Goods = req.GoodsId
	inventory.Stocks = req.Num

	if result := global.MySQLConn.Save(&inventory); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
		return &empty.Empty{}, result.Error
	}
	return &empty.Empty{}, nil
}

func (hadnler *InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inventory model.Inventory

	if result := global.MySQLConn.Where(&model.Inventory{Goods: req.GoodsId}).First(&inventory); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "没有库存信息")
	}
	return &proto.GoodsInvInfo{
		GoodsId: inventory.Goods,
		Num:     inventory.Stocks,
	}, nil
}

func (hadnler *InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*empty.Empty, error) {
	// todo 这里有事务问题，第一件商品扣减成功，但是第二件商品因为库存没有扣减

	tx := global.MySQLConn.Begin()

	for _, goodInfo := range req.GoodsInfo {
		var inventory model.Inventory
		for true {
			if result := global.MySQLConn.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inventory); result.RowsAffected == 0 {
				tx.Rollback()
				zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
				return nil, status.Errorf(codes.NotFound, "没有库存信息")
			}
			if inventory.Stocks < goodInfo.Num {
				tx.Rollback()
				zap.S().Errorw("goodsInventory.Stocks < goods.Num failed")
				return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
			}
			inventory.Stocks -= goodInfo.Num

			// 乐观锁：失败之后主动重新查询数据库再次更新库存，比较积极主动所以叫乐观锁，相比较于悲观锁在查询的时候把记录锁住更积极尝试，而不是坐吃等锁
			// 1. 更新的时候需要两个字段，商品ID、和版本号，更新之后版本号加1
			// 2. 因为更新方法是原子性的，所以不会导致并发问题
			// 3. 需要注意的是注意零值的更新，查询的时候不能使用tx查询，需要用global.MySQLConn
			if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version= ?", goodInfo.GoodsId, inventory.Version).Updates(model.Inventory{Stocks: inventory.Stocks, Version: inventory.Version + 1}); result.RowsAffected == 0 {
				zap.S().Info("库存扣减失败")
			} else {
				break
			}
		}
	}

	tx.Commit()

	return &empty.Empty{}, nil
}

func (hadnler *InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*empty.Empty, error) {
	tx := global.MySQLConn.Begin()

	for _, goods := range req.GoodsInfo {
		var goodsInventory model.Inventory
		if result := global.MySQLConn.Where(&model.Inventory{Goods: goods.GoodsId}).First(&goodsInventory); result.RowsAffected == 0 {
			tx.Rollback()
			zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
			return nil, status.Errorf(codes.NotFound, "没有库存信息")
		}

		goodsInventory.Stocks += goods.Num
		if result := tx.Save(&goodsInventory); result.Error != nil {
			tx.Rollback()
			zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
			return nil, status.Errorf(codes.ResourceExhausted, "归还库存失败")
		}
	}
	tx.Commit()
	return &empty.Empty{}, nil
}
