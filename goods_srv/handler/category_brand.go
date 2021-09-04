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

func (handler *GoodsServer) CategoryBrandList(ctx context.Context, req *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	var categoryBrands []model.GoodsCategoryBrand
	var rowCount int

	result := global.MySQLConn.Preload("Category").Preload("Brands").Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&categoryBrands)
	if result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Scopes failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "查询分类品牌列表失败")
	}
	rowCount = int(result.RowsAffected)

	categoryBrandResponses := make([]*proto.CategoryBrandResponse, rowCount)
	for i := 0; i < rowCount; i++ {
		categoryBrandResponse := &proto.CategoryBrandResponse{
			Id: categoryBrands[i].ID,
			Brand: &proto.BrandInfoResponse{
				Id:   categoryBrands[i].Brands.ID,
				Name: categoryBrands[i].Brands.Name,
				Logo: categoryBrands[i].Brands.Logo,
			},
			Category: &proto.CategoryInfoResponse{
				Id:             categoryBrands[i].Category.ID,
				Name:           categoryBrands[i].Category.Name,
				ParentCategory: categoryBrands[i].Category.ParentCategoryID,
				Level:          categoryBrands[i].Category.Level,
				IsTab:          categoryBrands[i].Category.IsTab,
			},
		}
		categoryBrandResponses[i] = categoryBrandResponse
	}

	return &proto.CategoryBrandListResponse{
		Total: int32(rowCount),
		Data:  categoryBrandResponses,
	}, nil
}

//通过category获取brands
func (handler *GoodsServer) GetCategoryBrandList(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	var category model.Category
	var rowCount int

	if result := global.MySQLConn.First(&category, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "分类不存在")
	}

	var categoryBrands []model.GoodsCategoryBrand
	result := global.MySQLConn.Where(&model.GoodsCategoryBrand{CategoryID: category.ID}).Preload("Brands").Find(&categoryBrands)
	if result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "查询分类品牌列表失败")
	}
	rowCount = int(result.RowsAffected)

	brandInfoResponses := make([]*proto.BrandInfoResponse, rowCount)
	for i := 0; i < rowCount; i++ {
		brandInfoResponse := &proto.BrandInfoResponse{
			Id:   categoryBrands[i].Brands.ID,
			Name: categoryBrands[i].Brands.Name,
			Logo: categoryBrands[i].Brands.Logo,
		}
		brandInfoResponses[i] = brandInfoResponse
	}

	return &proto.BrandListResponse{
		Total: int32(rowCount),
		Data:  brandInfoResponses,
	}, nil
}

func (handler *GoodsServer) CreateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	if result := global.MySQLConn.First(&model.Category{}, req.CategoryId); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "分类不存在")
	}

	if result := global.MySQLConn.First(&model.Brands{}, req.BrandId); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "品牌不存在")
	}

	categoryBrand := model.GoodsCategoryBrand{
		CategoryID: req.CategoryId,
		BrandsID:   req.BrandId,
	}
	if result := global.MySQLConn.Create(&categoryBrand); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Create failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "创建分类品牌失败")
	}

	return &proto.CategoryBrandResponse{Id: categoryBrand.CategoryID}, nil
}

func (handler *GoodsServer) DeleteCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*empty.Empty, error) {
	if result := global.MySQLConn.Where(&model.GoodsCategoryBrand{
		CategoryID: req.CategoryId,
		BrandsID:   req.BrandId,
	}).Delete(&model.GoodsCategoryBrand{}); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "分类品牌不存在")
	}
	return &empty.Empty{}, nil
}

func (handler *GoodsServer) UpdateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*empty.Empty, error) {
	if result := global.MySQLConn.First(&model.Category{}, req.CategoryId); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "分类不存在")
	}

	if result := global.MySQLConn.First(&model.Brands{}, req.BrandId); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "品牌不存在")
	}

	var categoryBrand model.GoodsCategoryBrand
	if result := global.MySQLConn.Where(&model.GoodsCategoryBrand{
		CategoryID: req.CategoryId,
		BrandsID:   req.BrandId,
	}).First(&categoryBrand); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "分类品牌不存在")
	}

	categoryBrand.BrandsID = req.BrandId
	categoryBrand.CategoryID = req.CategoryId
	if result := global.MySQLConn.Save(&categoryBrand); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "更新分类品牌失败")
	}

	return &empty.Empty{}, nil
}
