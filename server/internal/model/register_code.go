package model

import "time"

// RegisterCode 管理后台生成的注册码，无软删除（用 status 禁用）。
type RegisterCode struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"column:code;type:varchar(32);not null;uniqueIndex:uk_code" json:"code"`
	InviterID uint64    `gorm:"column:inviter_id;type:bigint;not null;default:0" json:"inviter_id"`
	Status    int8      `gorm:"column:status;type:tinyint;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (RegisterCode) TableName() string { return "register_codes" }
