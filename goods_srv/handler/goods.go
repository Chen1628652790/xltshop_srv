package handler

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xlt/shop_srv/goods_srv/global"
	"github.com/xlt/shop_srv/goods_srv/model"
	"github.com/xlt/shop_srv/goods_srv/proto"
)

type GoodsServer struct {
	proto.UnimplementedGoodsServer
}

func (handler *GoodsServer) GoodsList(ctx context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	mysqlDB := global.MySQLConn.Model(&model.Goods{})
	if req.KeyWords != "" {
		mysqlDB = mysqlDB.Where("name LIKE ?", req.KeyWords)
	}
	if req.IsHot {
		mysqlDB = mysqlDB.Where(&model.Goods{IsHot: true})
	}
	if req.IsNew {
		mysqlDB = mysqlDB.Where(&model.Goods{IsNew: true})
	}
	if req.PriceMin > 0 {
		mysqlDB = mysqlDB.Where("shop_price >= ?", req.PriceMin)
	}
	if req.PriceMax > 0 {
		mysqlDB = mysqlDB.Where("shop_price <= ?", req.PriceMax)
	}
	if req.Brand > 0 {
		mysqlDB = mysqlDB.Where(&model.Goods{BrandsID: req.Brand})
	}

	//通过category去查询商品
	var subQuery string
	if req.TopCategory > 0 {
		var category model.Category
		if result := global.MySQLConn.First(&category, req.TopCategory); result.RowsAffected == 0 {
			zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
			return nil, status.Errorf(codes.NotFound, "商品分类不存在")
		}

		if category.Level == 1 {
			subQuery = fmt.Sprintf("select id from category where parent_category_id in (select id from category WHERE parent_category_id=%d)", req.TopCategory)
		} else if category.Level == 2 {
			subQuery = fmt.Sprintf("select id from category WHERE parent_category_id=%d", req.TopCategory)
		} else if category.Level == 3 {
			subQuery = fmt.Sprintf("select id from category WHERE id=%d", req.TopCategory)
		}
		mysqlDB = mysqlDB.Where(fmt.Sprintf("category_id in (%s)", subQuery))
	}

	var goods []model.Goods
	var rowCount int
	var tmpCount int64

	mysqlDB.Count(&tmpCount)
	rowCount = int(tmpCount)
	result := mysqlDB.Preload("Category").Preload("Brands").Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&goods)
	if result.Error != nil {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "查询商品失败")
	}

	goodsInfoResponses := make([]*proto.GoodsInfoResponse, rowCount)
	for i := 0; i < rowCount; i++ {
		goodsInfoResponse := ModelToResponse(goods[i])
		goodsInfoResponses[i] = &goodsInfoResponse
	}

	return &proto.GoodsListResponse{
		Total: int32(rowCount),
		Data:  goodsInfoResponses,
	}, nil
}

func (handler *GoodsServer) BatchGetGoods(ctx context.Context, req *proto.BatchGoodsIdInfo) (*proto.GoodsListResponse, error) {
	var goods []model.Goods
	var rowCount int

	if result := global.MySQLConn.Find(&goods, req.Id); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "查询商品失败")
	}

	goodsInfoResponses := make([]*proto.GoodsInfoResponse, rowCount)
	for i := 0; i < rowCount; i++ {
		goodsInfoResponse := ModelToResponse(goods[i])
		goodsInfoResponses[i] = &goodsInfoResponse
	}

	return &proto.GoodsListResponse{
		Total: int32(rowCount),
		Data:  goodsInfoResponses,
	}, nil
}

func (handler *GoodsServer) GetGoodsDetail(ctx context.Context, req *proto.GoodInfoRequest) (*proto.GoodsInfoResponse, error) {
	var goods model.Goods

	if result := global.MySQLConn.Find(&goods, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	goodsInfoRsp := ModelToResponse(goods)
	return &goodsInfoRsp, nil
}

func (handler *GoodsServer) CreateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	var category model.Category
	if result := global.MySQLConn.Find(&category, req.CategoryId); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}

	var brand model.Brands
	if result := global.MySQLConn.Find(&brand); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}

	goods := model.Goods{
		CategoryID:      category.ID,
		BrandsID:        brand.ID,
		OnSale:          req.OnSale,
		ShipFree:        req.ShipFree,
		IsNew:           req.IsNew,
		IsHot:           req.IsHot,
		Name:            req.Name,
		GoodsSn:         req.GoodsSn,
		MarketPrice:     req.MarketPrice,
		ShopPrice:       req.ShopPrice,
		GoodsBrief:      req.GoodsBrief,
		Images:          req.Images,
		DescImages:      req.DescImages,
		GoodsFrontImage: req.GoodsFrontImage,
	}
	if result := global.MySQLConn.Create(&goods); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Create failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "创建商品失败")
	}

	return &proto.GoodsInfoResponse{Id: goods.ID}, nil
}

func (handler *GoodsServer) DeleteGoods(ctx context.Context, req *proto.DeleteGoodsInfo) (*empty.Empty, error) {
	if result := global.MySQLConn.Delete(&model.Goods{}, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Delete failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "商品不存在")
	}
	return &empty.Empty{}, nil
}

func (handler *GoodsServer) UpdateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*empty.Empty, error) {
	var goods model.Goods
	if result := global.MySQLConn.Find(&goods, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "商品不存在")
	}
	if result := global.MySQLConn.Find(&model.Category{}, req.CategoryId); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}
	if result := global.MySQLConn.Find(&model.Brands{}, req.BrandId); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}

	goods = model.Goods{
		CategoryID:      req.CategoryId,
		BrandsID:        req.BrandId,
		OnSale:          req.OnSale,
		ShipFree:        req.ShipFree,
		IsNew:           req.IsNew,
		IsHot:           req.IsHot,
		Name:            req.Name,
		GoodsSn:         req.GoodsSn,
		MarketPrice:     req.MarketPrice,
		ShopPrice:       req.ShopPrice,
		GoodsBrief:      req.GoodsBrief,
		Images:          req.Images,
		DescImages:      req.DescImages,
		GoodsFrontImage: req.GoodsFrontImage,
	}
	if result := global.MySQLConn.Save(&goods); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Create failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "更新商品失败")
	}

	return &empty.Empty{}, nil
}

func ModelToResponse(goods model.Goods) proto.GoodsInfoResponse {
	return proto.GoodsInfoResponse{
		Id:              goods.ID,
		CategoryId:      goods.CategoryID,
		Name:            goods.Name,
		GoodsSn:         goods.GoodsSn,
		ClickNum:        goods.ClickNum,
		SoldNum:         goods.SolidNum,
		FavNum:          goods.FavNum,
		MarketPrice:     goods.MarketPrice,
		ShopPrice:       goods.ShopPrice,
		GoodsBrief:      goods.GoodsBrief,
		ShipFree:        goods.ShipFree,
		GoodsFrontImage: goods.GoodsFrontImage,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
		OnSale:          goods.OnSale,
		DescImages:      goods.DescImages,
		Images:          goods.Images,
		Category: &proto.CategoryBriefInfoResponse{
			Id:   goods.Category.ID,
			Name: goods.Category.Name,
		},
		Brand: &proto.BrandInfoResponse{
			Id:   goods.Brands.ID,
			Name: goods.Brands.Name,
			Logo: goods.Brands.Logo,
		},
	}
}
