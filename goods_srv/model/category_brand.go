package model

type GoodsCategoryBrand struct {
	BaseModel
	CategoryID int32 `gorm:"type:int;index_idx_category_brand,unique"`
	Category   Category

	BrandsID int32 `gorm:"type:int;index_idx_category_brand,unique"`
	Brands   Brands
}

func (g *GoodsCategoryBrand) TableName() string {
	return "goodscategorybrand"
}
