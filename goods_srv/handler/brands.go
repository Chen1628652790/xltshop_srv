package handler

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/xlt/shop_srv/goods_srv/global"
	"github.com/xlt/shop_srv/goods_srv/model"
	"github.com/xlt/shop_srv/goods_srv/proto"
)

func (handler *GoodsServer) BrandList(ctx context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	var brands []model.Brands
	var rowCount int

	result := global.MySQLConn.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&brands)
	if result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.InvalidArgument, "查询品牌列表失败")
	}
	rowCount = int(result.RowsAffected)

	brandInfoRsp := make([]*proto.BrandInfoResponse, rowCount)
	for i := 0; i < rowCount; i++ {
		brand := &proto.BrandInfoResponse{
			Id:   brands[i].ID,
			Name: brands[i].Name,
			Logo: brands[i].Logo,
		}
		brandInfoRsp[i] = brand
	}

	return &proto.BrandListResponse{
		Total: int32(rowCount),
		Data:  brandInfoRsp,
	}, nil
}

func (handler *GoodsServer) CreateBrand(ctx context.Context, req *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	if result := global.MySQLConn.First(&model.Brands{}); result.RowsAffected != 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.InvalidArgument, "品牌已存在")
	}

	brand := model.Brands{
		Name: req.Name,
		Logo: req.Logo,
	}
	if result := global.MySQLConn.Create(&brand); result.Error != nil {
		return nil, status.Errorf(codes.Internal, "创建品牌失败")
	}

	return &proto.BrandInfoResponse{Id: brand.ID}, nil
}

func (handler *GoodsServer) DeleteBrand(ctx context.Context, req *proto.BrandRequest) (*empty.Empty, error) {
	if result := global.MySQLConn.Delete(&model.Brands{}, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Delete failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}
	return &empty.Empty{}, nil
}

func (handler *GoodsServer) UpdateBrand(ctx context.Context, req *proto.BrandRequest) (*empty.Empty, error) {
	if result := global.MySQLConn.First(&model.Brands{}, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}

	brand := model.Brands{}
	if req.Name != "" {
		brand.Name = req.Name
	}
	if req.Logo != "" {
		brand.Logo = req.Logo
	}

	if result := global.MySQLConn.Save(&brand); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "更新品牌失败")
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
