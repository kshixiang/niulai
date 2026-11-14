package model

// 产品状态
const (
	ProductStatusOffline int8 = 0
	ProductStatusOnline  int8 = 1
)

// Product 充值产品（话费/流量/油卡等）
type Product struct {
	BaseModel
	Name      string `gorm:"column:name;type:varchar(128);not null;comment:产品名称" json:"name"`
	Category  string `gorm:"column:category;type:varchar(32);not null;index;comment:产品类目" json:"category"`
	FaceValue Money  `gorm:"column:face_value;type:decimal(18,2);default:0;not null;comment:面值" json:"face_value"`
	SalePrice Money  `gorm:"column:sale_price;type:decimal(18,2);default:0;not null;comment:售价" json:"sale_price"`
	CostPrice Money  `gorm:"column:cost_price;type:decimal(18,2);default:0;not null;comment:成本价" json:"cost_price"`
	ChannelID uint64 `gorm:"column:channel_id;type:bigint;not null;index;comment:供货渠道ID" json:"channel_id"`
	Stock     int64  `gorm:"column:stock;type:bigint;default:-1;not null;comment:库存 -1表示不限" json:"stock"`
	Sort      int    `gorm:"column:sort;type:int;default:0;not null;comment:排序" json:"sort"`
	Status    int8   `gorm:"column:status;type:tinyint;default:1;not null;index;comment:状态 0-下架 1-上架" json:"status"`
}

func (Product) TableName() string { return "products" }
