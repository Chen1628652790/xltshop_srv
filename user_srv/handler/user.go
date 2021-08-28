package handler

import (
	"context"

	"github.com/xlt/shop_srv/user_srv/global"
	"github.com/xlt/shop_srv/user_srv/model"
	"github.com/xlt/shop_srv/user_srv/proto"
	"gorm.io/gorm"
)

type UserServer struct {
}

func (h *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {
	var users []model.User
	var count int64

	result := global.DB.Model(&model.User{}).Count(&count)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.UserListResponse{
		Total: int32(count),
	}

	global.DB.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&users)

	for _, user := range users {
		userInfoRsp := ModelToResponse(user)
		rsp.Data = append(rsp.Data, userInfoRsp)
	}

	return rsp, nil
}

func ModelToResponse(user model.User) *proto.UserInfoResponse {
	userInfoRsp := proto.UserInfoResponse{
		Id:       user.ID,
		PassWord: user.Password,
		Mobile:   user.Mobile,
		NickName: user.NickName,
		Birthday: 0,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}
	if user.Birthday != nil {
		userInfoRsp.Birthday = uint64(user.Birthday.Unix())
	}

	return &userInfoRsp
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
