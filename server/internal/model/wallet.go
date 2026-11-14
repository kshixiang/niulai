package model

import "time"

// Wallet 资金账户。账本类数据不做软删除，因此不内嵌 BaseModel。
type Wallet struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt     time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	UserID        uint64    `gorm:"column:user_id;type:bigint;not null;uniqueIndex;comment:用户ID(关联users.id)" json:"user_id"`
	Balance       Money     `gorm:"column:balance;type:decimal(10,2);not null;default:0;comment:可用余额" json:"balance"`
	FrozenBalance Money     `gorm:"column:frozen_balance;type:decimal(10,2);not null;default:0;comment:冻结余额" json:"frozen_balance"`
	Version       int       `gorm:"column:version;type:int;not null;default:1;comment:乐观锁版本号" json:"version"`
}

func (Wallet) TableName() string { return "wallets" }
