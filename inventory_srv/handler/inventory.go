package handler

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm/clause"

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

	for _, goods := range req.GoodsInfo {
		var goodsInventory model.Inventory

		// select * from inventory for update --- 悲观锁
		// 使用前提：需要开启事务，这样就不会自动提交，可以在事务中进行操作然后手动提交
		// 1. 悲观锁由来：因为每次操作都觉得需要的数据可能会被其它线程操作（老是觉得有问题比较悲观），需要自己先抢到锁才可以进行操作
		// 2. 悲观锁的使用：在查询语句时候悲观锁，其它线程进来抢不到锁会阻塞，事务提交之后其它线程才能操作，保证数据一致性
		// 3. 悲观锁粒度升级：悲观锁在使用索引查询的时候是行锁，锁定一行记录。如果没有使用索引查询的话会导致锁表，行锁升级为表锁。
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods: goods.GoodsId}).First(&goodsInventory); result.RowsAffected == 0 {
			tx.Rollback()
			zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
			return nil, status.Errorf(codes.NotFound, "没有库存信息")
		}
		if goodsInventory.Stocks < goods.Num {
			tx.Rollback()
			zap.S().Errorw("goodsInventory.Stocks < goods.Num failed")
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}

		goodsInventory.Stocks -= goods.Num
		if result := tx.Save(&goodsInventory); result.Error != nil {
			tx.Rollback()
			zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
			return nil, status.Errorf(codes.ResourceExhausted, "扣减库存失败")
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
