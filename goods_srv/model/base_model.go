package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type BaseModel struct {
	ID        int32     `gorm:"primarykey;type:int" json:"id"` //为什么使用int32， bigint
	CreatedAt time.Time `gorm:"column:add_time" json:"-"`
	UpdatedAt time.Time `gorm:"column:update_time" json:"-"`
	//DeletedAt gorm.DeletedAt `json:"-"`
	IsDeleted bool `json:"-"`
}

type GormList []string

func (g GormList) Value() (value driver.Value, err error) {
	return json.Marshal(g)
}

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)
}
