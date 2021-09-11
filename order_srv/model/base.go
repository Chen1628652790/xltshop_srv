package model

import (
	"time"
)

type BaseModel struct {
	ID        int32     `gorm:"primarykey;type:int" json:"id"` //为什么使用int32， bigint
	CreatedAt time.Time `gorm:"column:add_time" json:"-"`
	UpdatedAt time.Time `gorm:"column:update_time" json:"-"`
	IsDeleted bool      `json:"-"`
}
