package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 通用基础模型。
// 所有业务表复用该结构，自带自增主键、创建/更新时间以及软删除支持。
// 使用 uint64 而非 int，避免数据量逼近 int 上限；DeletedAt 交由 GORM 自动维护软删除。
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;not null;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间(软删除)" json:"-"`
}
