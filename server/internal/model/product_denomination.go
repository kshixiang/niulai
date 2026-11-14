package model

// ProductDenomination 商品价格类别(面值)。
// 一个商品可配置多个面值，共用商品级折扣；折后价不落库，由 面值×折扣÷10 实时计算。
type ProductDenomination struct {
	BaseModel
	ProductID uint64 `gorm:"column:product_id;type:bigint;not null;default:0;index:idx_product_id;comment:商品ID" json:"product_id"`
	Name      string `gorm:"column:name;type:varchar(64);not null;default:'';comment:价格类别名称 可选" json:"name"`
	FaceValue Money  `gorm:"column:face_value;type:decimal(10,2);not null;default:0;comment:面值/原价" json:"face_value"`
	Sort      int    `gorm:"column:sort;type:int;not null;default:0;comment:展示排序 升序" json:"sort"`
}

func (ProductDenomination) TableName() string { return "product_denominations" }
