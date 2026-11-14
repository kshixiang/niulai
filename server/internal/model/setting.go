package model

import "time"

// Setting 系统设置（key-value），无软删除。
type Setting struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ConfigKey   string    `gorm:"column:config_key;type:varchar(64);not null;uniqueIndex:uk_config_key" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;type:varchar(255);not null;default:''" json:"config_value"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Setting) TableName() string { return "settings" }
