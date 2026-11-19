package model

import "time"

const (
	RechargePending  int8 = 0 // 待核实
	RechargeApproved int8 = 1 // 已到账
	RechargeRejected int8 = 2 // 已驳回
)

// Recharge 保留申报金额与收款地址快照；交易哈希全局唯一，审核后不允许删除或重审。
type Recharge struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt      time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null" json:"updated_at"`
	UserID         uint64     `gorm:"not null;index:idx_recharge_user_created,priority:1" json:"user_id"`
	Address        string     `gorm:"type:varchar(34);not null" json:"address"`
	TxHash         string     `gorm:"type:char(64);not null;uniqueIndex:uk_recharge_tx_hash" json:"tx_hash"`
	ReportedAmount string     `gorm:"type:varchar(32);not null" json:"reported_amount"`
	Amount         Money      `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status         int8       `gorm:"type:tinyint;not null;default:0;index:idx_recharge_status" json:"status"`
	Remark         string     `gorm:"type:varchar(255);not null;default:''" json:"remark"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
}

func (Recharge) TableName() string { return "recharges" }
