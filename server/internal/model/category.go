package model

import "time"

// ProductCategory 的 Code 是固定关联标识；改名只更新 Name，不影响话费业务规则。
type ProductCategory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"type:varchar(32);not null;uniqueIndex:uk_product_category_code" json:"code"`
	Name      string    `gorm:"type:varchar(32);not null;uniqueIndex:uk_product_category_name" json:"name"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (ProductCategory) TableName() string { return "product_categories" }
