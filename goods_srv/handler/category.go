package handler

import (
	"context"
	"encoding/json"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xlt/shop_srv/goods_srv/global"
	"github.com/xlt/shop_srv/goods_srv/model"
	"github.com/xlt/shop_srv/goods_srv/proto"
)

func (handler *GoodsServer) GetAllCategorysList(ctx context.Context, req *empty.Empty) (*proto.CategoryListResponse, error) {
	var categorys []model.Category

	if result := global.MySQLConn.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categorys); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error)
		return nil, status.Errorf(codes.Internal, "查询分类列表失败")
	}

	categorysJSON, err := json.Marshal(&categorys)
	if err != nil {
		zap.S().Errorw("json.Marshal failed", "msg", err.Error())
		return nil, status.Errorf(codes.Internal, "查询分类列表失败")
	}

	return &proto.CategoryListResponse{JsonData: string(categorysJSON)}, nil
}

//获取子分类
func (handler *GoodsServer) GetSubCategory(ctx context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	var category model.Category
	var rowCount int

	if result := global.MySQLConn.Find(&category, req.Id); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Find failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "不存在此分类")
	}

	categoryInfoRsp := &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		ParentCategory: category.ParentCategoryID,
		Level:          category.Level,
		IsTab:          category.IsTab,
	}

	var subCategory []model.Category
	preloads := ""
	if category.ParentCategoryID == 1 {
		preloads = "SubCategory.SubCategory"
	} else if category.ParentCategoryID == 2 {
		preloads = "SubCategory"
	}
	result := global.MySQLConn.Where(&model.Category{
		ParentCategoryID: category.ID,
	}).Preload(preloads).Find(&subCategory)
	if result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Where failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "查询子分类失败")
	}
	rowCount = int(result.RowsAffected)

	subCategoryInfoRsp := make([]*proto.CategoryInfoResponse, rowCount)
	for i := 0; i < rowCount; i++ {
		value := &proto.CategoryInfoResponse{
			Id:             subCategory[i].ID,
			Name:           subCategory[i].Name,
			ParentCategory: subCategory[i].ParentCategoryID,
			Level:          subCategory[i].Level,
			IsTab:          subCategory[i].IsTab,
		}
		subCategoryInfoRsp[i] = value
	}

	return &proto.SubCategoryListResponse{
		Total:        int32(rowCount),
		Info:         categoryInfoRsp,
		SubCategorys: subCategoryInfoRsp,
	}, nil
}

func (handler *GoodsServer) CreateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	category := model.Category{
		Name:             req.Name,
		ParentCategoryID: req.ParentCategory,
		Level:            req.Level,
		IsTab:            req.IsTab,
	}
	if result := global.MySQLConn.Create(&category); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Create failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "创建分类失败")
	}

	return &proto.CategoryInfoResponse{Id: category.ID}, nil
}

func (handler *GoodsServer) DeleteCategory(ctx context.Context, req *proto.DeleteCategoryRequest) (*empty.Empty, error) {
	if result := global.MySQLConn.Delete(&model.Category{}, req.Id); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.Delete failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "分类不存在")
	}
	return &empty.Empty{}, nil
}

func (handler *GoodsServer) UpdateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*empty.Empty, error) {
	var category model.Category
	if result := global.MySQLConn.First(&category); result.RowsAffected == 0 {
		zap.S().Errorw("global.MySQLConn.First failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "分类不存在")
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Level != 0 {
		category.Level = req.Level
	}
	if req.IsTab != false {
		category.IsTab = req.IsTab
	}
	if req.ParentCategory != 0 {
		category.ParentCategoryID = req.ParentCategory
	}

	if result := global.MySQLConn.Save(&category); result.Error != nil {
		zap.S().Errorw("global.MySQLConn.Save failed", "msg", result.Error.Error())
		return nil, status.Errorf(codes.Internal, "更新分类失败")
	}
	return &empty.Empty{}, nil
}
