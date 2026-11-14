package model

import "time"

// 分润状态
const (
	CommissionStatusPending int8 = 0 // 待结算
	CommissionStatusSettled int8 = 1 // 已结算
)

// CommissionRecord 分润流水。账本类数据不做软删除，因此不内嵌 BaseModel。
type CommissionRecord struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	OrderID   uint64    `gorm:"column:order_id;type:bigint;not null;index;comment:订单ID(关联orders.id)" json:"order_id"`
	UserID    uint64    `gorm:"column:user_id;type:bigint;not null;index:idx_user_status,priority:1;comment:获得分润的代理(关联users.id)" json:"user_id"`
	Amount    Money     `gorm:"column:amount;type:decimal(10,2);not null;default:0;comment:分润金额" json:"amount"`
	Level     int8      `gorm:"column:level;type:tinyint;not null;default:1;index;comment:分润层级" json:"level"`
	Status    int8      `gorm:"column:status;type:tinyint;not null;default:0;index:idx_user_status,priority:2;comment:0待结算 1已结算" json:"status"`
}

func (CommissionRecord) TableName() string { return "commission_records" }
