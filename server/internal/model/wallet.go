package model

// 钱包拥有者类型
const (
	WalletOwnerUser  int8 = 1
	WalletOwnerAgent int8 = 2
)

// Wallet 资金账户（用户与代理共用，通过 owner_type 区分）。
// Version 字段用于乐观锁，配合 GORM 的 Clause 实现并发扣款安全。
type Wallet struct {
	BaseModel
	OwnerID       uint64 `gorm:"column:owner_id;type:bigint;not null;index;comment:拥有者ID" json:"owner_id"`
	OwnerType     int8   `gorm:"column:owner_type;type:tinyint;not null;index;comment:拥有者类型 1-用户 2-代理" json:"owner_type"`
	Balance       Money  `gorm:"column:balance;type:decimal(18,2);default:0;not null;comment:可用余额" json:"balance"`
	FrozenAmount  Money  `gorm:"column:frozen_amount;type:decimal(18,2);default:0;not null;comment:冻结金额" json:"frozen_amount"`
	TotalIncome   Money  `gorm:"column:total_income;type:decimal(18,2);default:0;not null;comment:累计收入" json:"total_income"`
	TotalWithdraw Money  `gorm:"column:total_withdraw;type:decimal(18,2);default:0;not null;comment:累计提现" json:"total_withdraw"`
	Version       int64  `gorm:"column:version;type:bigint;default:0;not null;comment:乐观锁版本号" json:"version"`
}

func (Wallet) TableName() string { return "wallets" }
