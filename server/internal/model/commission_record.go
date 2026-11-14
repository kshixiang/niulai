package model

import "time"

// 佣金记录状态
const (
	CommissionStatusPending   int8 = 0 // 待结算
	CommissionStatusSettled   int8 = 1 // 已结算
	CommissionStatusWithdrawn int8 = 2 // 已提现
	CommissionStatusCanceled  int8 = 3 // 已取消
)

// CommissionRecord 佣金分润记录（多级分销的结算明细）
type CommissionRecord struct {
	BaseModel
	OrderID     uint64     `gorm:"column:order_id;type:bigint;not null;index;comment:关联订单ID" json:"order_id"`
	OrderNo     string     `gorm:"column:order_no;type:varchar(32);not null;index;comment:订单号" json:"order_no"`
	AgentID     uint64     `gorm:"column:agent_id;type:bigint;not null;index;comment:受益代理ID" json:"agent_id"`
	FromAgentID uint64     `gorm:"column:from_agent_id;type:bigint;default:0;not null;comment:直接上级代理ID" json:"from_agent_id"`
	Level       int8       `gorm:"column:level;type:tinyint;default:1;not null;comment:分润层级" json:"level"`
	Amount      Money      `gorm:"column:amount;type:decimal(18,2);default:0;not null;comment:佣金金额" json:"amount"`
	Rate        Money      `gorm:"column:rate;type:decimal(8,4);default:0;not null;comment:分润比例" json:"rate"`
	Status      int8       `gorm:"column:status;type:tinyint;default:0;not null;index;comment:状态 0-待结算 1-已结算 2-已提现 3-已取消" json:"status"`
	SettledAt   *time.Time `gorm:"column:settled_at;comment:结算时间" json:"settled_at"`
}

func (CommissionRecord) TableName() string { return "commission_records" }
