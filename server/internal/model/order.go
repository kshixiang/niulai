package model

import "time"

// 订单状态
const (
	OrderStatusPending    int8 = 0 // 待支付
	OrderStatusPaid       int8 = 1 // 已支付
	OrderStatusRecharging int8 = 2 // 充值中
	OrderStatusSuccess    int8 = 3 // 充值成功
	OrderStatusFailed     int8 = 4 // 充值失败
	OrderStatusRefunded   int8 = 5 // 已退款
)

// Order 充值订单
type Order struct {
	BaseModel
	OrderNo        string     `gorm:"column:order_no;type:varchar(32);uniqueIndex;not null;comment:订单号" json:"order_no"`
	UserID         uint64     `gorm:"column:user_id;type:bigint;not null;index;comment:下单用户ID" json:"user_id"`
	ProductID      uint64     `gorm:"column:product_id;type:bigint;not null;index;comment:产品ID" json:"product_id"`
	ChannelID      uint64     `gorm:"column:channel_id;type:bigint;not null;index;comment:渠道ID" json:"channel_id"`
	AgentID        uint64     `gorm:"column:agent_id;type:bigint;default:0;not null;index;comment:成交代理ID" json:"agent_id"`
	Amount         Money      `gorm:"column:amount;type:decimal(18,2);default:0;not null;comment:订单金额" json:"amount"`
	CostAmount     Money      `gorm:"column:cost_amount;type:decimal(18,2);default:0;not null;comment:渠道成本" json:"cost_amount"`
	AgentCommission Money     `gorm:"column:agent_commission;type:decimal(18,2);default:0;not null;comment:代理佣金" json:"agent_commission"`
	Status         int8       `gorm:"column:status;type:tinyint;default:0;not null;index;comment:订单状态" json:"status"`
	PayType        string     `gorm:"column:pay_type;type:varchar(16);default:'';comment:支付方式" json:"pay_type"`
	PayTime        *time.Time `gorm:"column:pay_time;comment:支付时间" json:"pay_time"`
	ChannelOrderNo string     `gorm:"column:channel_order_no;type:varchar(64);default:'';comment:渠道侧订单号" json:"channel_order_no"`
	Remark         string     `gorm:"column:remark;type:varchar(255);default:'';comment:备注" json:"remark"`
}

func (Order) TableName() string { return "orders" }
