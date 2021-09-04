package handler

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xlt/shop_srv/goods_srv/global"
	"github.com/xlt/shop_srv/goods_srv/model"
	"github.com/xlt/shop_srv/goods_srv/proto"
)

func (hander *GoodsServer) BannerList(ctx context.Context, req *empty.Empty) (*proto.BannerListResponse, error) {
	var banners []model.Banner
	var rowCount int

	result := global.MySQLConn.Find(&banners)
	if result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "查询轮播图列表失败")
	}
	rowCount = int(result.RowsAffected)

	bannerRsp := make([]*proto.BannerResponse, rowCount)
	for i := 0; i < rowCount; i++ {
		banner := &proto.BannerResponse{
			Id:    banners[i].ID,
			Index: banners[i].Index,
			Image: banners[i].Image,
			Url:   banners[i].Url,
		}
		bannerRsp[i] = banner
	}

	return &proto.BannerListResponse{
		Total: int32(rowCount),
		Data:  bannerRsp,
	}, nil
}

func (hander *GoodsServer) CreateBanner(ctx context.Context, req *proto.BannerRequest) (*proto.BannerResponse, error) {
	banner := model.Banner{
		Image: req.Image,
		Url:   req.Url,
		Index: req.Index,
	}

	if result := global.MySQLConn.Create(&banner); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Create failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "创建轮播图失败")
	}
	return &proto.BannerResponse{Id: banner.ID}, nil
}

func (hander *GoodsServer) DeleteBanner(ctx context.Context, req *proto.BannerRequest) (*empty.Empty, error) {
	if result := global.MySQLConn.Delete(&model.Banner{}, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Delete failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "轮播图不存在")
	}
	return &empty.Empty{}, nil
}

func (hander *GoodsServer) UpdateBanner(ctx context.Context, req *proto.BannerRequest) (*empty.Empty, error) {
	banner := model.Banner{}

	if result := global.MySQLConn.First(&model.Banner{}, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "轮播图不存在")
	}

	if req.Url != "" {
		banner.Url = req.Url
	}
	if req.Image != "" {
		banner.Image = req.Url
	}
	if req.Index != 0 {
		banner.Index = req.Index
	}
	if result := global.MySQLConn.Save(&banner); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "更新轮播图失败")
	}

	return &empty.Empty{}, nil
}
