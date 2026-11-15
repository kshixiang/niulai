package model

import "time"

// 订单状态
const (
	OrderStatusPending    int8 = 0 // 待支付
	OrderStatusPaid       int8 = 1 // 已支付
	OrderStatusRecharging int8 = 2 // 充值中
	OrderStatusSuccess    int8 = 3 // 成功
	OrderStatusFailed     int8 = 4 // 失败
	OrderStatusRefunded   int8 = 5 // 已退款
)

// Order 充值订单
type Order struct {
	BaseModel
	OrderNo        string     `gorm:"column:order_no;type:varchar(64);uniqueIndex;not null;default:'';comment:唯一订单号" json:"order_no"`
	UserID         uint64     `gorm:"column:user_id;type:bigint;not null;index:idx_user_status,priority:1;comment:用户ID(关联users.id)" json:"user_id"`
	ProductID      uint64     `gorm:"column:product_id;type:bigint;not null;index;comment:商品ID(关联products.id)" json:"product_id"`
	ChannelID      uint64     `gorm:"column:channel_id;type:bigint;not null;default:0;index;comment:渠道ID(关联channels.id)" json:"channel_id"`
	Amount         Money      `gorm:"column:amount;type:decimal(10,2);not null;default:0;comment:实付金额" json:"amount"`
	ReceivedAmount Money      `gorm:"column:received_amount;type:decimal(10,2);not null;default:0;comment:已到账金额" json:"received_amount"`
	CostAmount     Money      `gorm:"column:cost_amount;type:decimal(10,2);not null;default:0;comment:渠道成本" json:"cost_amount"`
	Profit         Money      `gorm:"column:profit;type:decimal(10,2);not null;default:0;comment:利润" json:"profit"`
	Status         int8       `gorm:"column:status;type:tinyint;not null;default:0;index:idx_user_status,priority:2;comment:0待支付 1已支付 2充值中 3成功 4失败 5已退款" json:"status"`
	Account        string     `gorm:"column:account;type:varchar(128);not null;default:'';index;comment:充值账号" json:"account"`
	FormData       string     `gorm:"column:form_data;type:text;comment:下单填写项快照(JSON数组)" json:"-"`
	CallbackAt     *time.Time `gorm:"column:callback_at;comment:回调时间" json:"callback_at"`
}

func (Order) TableName() string { return "orders" }
