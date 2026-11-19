package model

import "time"

// 资金流水类型
const (
	WalletTxOrderFreeze    int8 = 1 // 下单冻结
	WalletTxDeduct         int8 = 2 // 成功扣款
	WalletTxUnfreezeFailed int8 = 3 // 失败解冻
	WalletTxRefund         int8 = 4 // 退款归还
	WalletTxRecharge       int8 = 5 // 充值到账
)

// WalletTransaction 资金流水。账本类数据不做软删除，因此不内嵌 BaseModel。
type WalletTransaction struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt     time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	UserID        uint64    `gorm:"column:user_id;type:bigint;not null;index:idx_user_created,priority:1;comment:用户ID(关联users.id)" json:"user_id"`
	OrderID       uint64    `gorm:"column:order_id;type:bigint;not null;default:0;index;comment:订单ID(关联orders.id)" json:"order_id"`
	Type          int8      `gorm:"column:type;type:tinyint;not null;comment:类型 1下单冻结 2成功扣款 3失败解冻 4退款归还 5充值到账" json:"type"`
	Amount        Money     `gorm:"column:amount;type:decimal(10,2);not null;default:0;comment:变动USDT金额(恒为正)" json:"amount"`
	BalanceBefore Money     `gorm:"column:balance_before;type:decimal(10,2);not null;default:0;comment:变动前可用余额" json:"balance_before"`
	BalanceAfter  Money     `gorm:"column:balance_after;type:decimal(10,2);not null;default:0;comment:变动后可用余额" json:"balance_after"`
	FrozenBefore  Money     `gorm:"column:frozen_before;type:decimal(10,2);not null;default:0;comment:变动前冻结余额" json:"frozen_before"`
	FrozenAfter   Money     `gorm:"column:frozen_after;type:decimal(10,2);not null;default:0;comment:变动后冻结余额" json:"frozen_after"`
	Ref           string    `gorm:"column:ref;type:varchar(64);not null;default:'';comment:业务号(订单号)" json:"ref"`
	Remark        string    `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
}

func (WalletTransaction) TableName() string { return "wallet_transactions" }
