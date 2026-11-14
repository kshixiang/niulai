package model

// ProductOrderField 商品下单填写项配置。
// options 存 JSON 数组字符串（如 ["移动","联通","电信"]），由 service 解析后对外提供数组。
type ProductOrderField struct {
	BaseModel
	ProductID   uint64 `gorm:"column:product_id;type:bigint;not null;default:0;uniqueIndex:uk_product_field,priority:1;comment:商品ID" json:"product_id"`
	FieldKey    string `gorm:"column:field_key;type:varchar(64);not null;default:'';uniqueIndex:uk_product_field,priority:2;comment:字段英文标识" json:"field_key"`
	Label       string `gorm:"column:label;type:varchar(128);not null;default:'';comment:中文名" json:"label"`
	Type        string `gorm:"column:type;type:varchar(16);not null;default:'text';comment:text/number/select" json:"type"`
	Required    int8   `gorm:"column:required;type:tinyint;not null;default:0;comment:1必填 0选填" json:"required"`
	Placeholder string `gorm:"column:placeholder;type:varchar(255);not null;default:'';comment:输入提示语" json:"placeholder"`
	Options     string `gorm:"column:options;type:text;comment:select选项 JSON数组字符串" json:"-"`
	Sort        int    `gorm:"column:sort;type:int;not null;default:0;comment:展示排序 升序" json:"sort"`
}

func (ProductOrderField) TableName() string { return "product_order_fields" }
