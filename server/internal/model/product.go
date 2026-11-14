package model

// 商品状态
const (
	ProductStatusOffline int8 = 0
	ProductStatusOnline  int8 = 1
)

// 商品类型白名单
const (
	ProductTypePhone = "话费"
	ProductTypeCard  = "卡券"
)

// Product 充值商品（话费 / 卡券 / 会员等）
type Product struct {
	BaseModel
	Name          string `gorm:"column:name;type:varchar(128);not null;default:'';comment:商品名称" json:"name"`
	Type          string `gorm:"column:type;type:varchar(32);not null;default:'';index:idx_type_status,priority:1;comment:类型 话费/卡券/会员" json:"type"`
	CoverImage    string `gorm:"column:cover_image;type:varchar(512);not null;default:'';comment:封面图" json:"cover_image"`
	SalePrice     Money  `gorm:"column:sale_price;type:decimal(10,2);not null;default:0;comment:销售价" json:"sale_price"`
	CostPrice     Money  `gorm:"column:cost_price;type:decimal(10,2);not null;default:0;comment:成本价" json:"cost_price"`
	DiscountLabel string `gorm:"column:discount_label;type:varchar(32);not null;default:'';comment:折扣标签 如7.8折" json:"discount_label"`
	Description   string `gorm:"column:description;type:varchar(512);not null;default:'';comment:提示语/描述" json:"description"`
	ChannelID     uint64 `gorm:"column:channel_id;type:bigint;not null;default:0;comment:绑定的上游渠道ID" json:"channel_id"`
	Status        int8   `gorm:"column:status;type:tinyint;not null;default:0;index:idx_type_status,priority:2;comment:上下架状态 1上架 0下架" json:"status"`
}

func (Product) TableName() string { return "products" }
