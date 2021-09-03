package handler

import (
	"context"

	"go.uber.org/zap"
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
		return nil, result.Error
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
