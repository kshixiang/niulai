package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"niulai/internal/model"
)

// SettingRegisterCodeRequired 注册码是否强制填写的配置键。
const SettingRegisterCodeRequired = "register_code_required"

// SettingService 系统设置读写。
type SettingService struct {
	db *gorm.DB
}

func NewSettingService(db *gorm.DB) *SettingService {
	return &SettingService{db: db}
}

// IsRegisterCodeRequired 读取开关，记录缺失时默认关闭。
func (s *SettingService) IsRegisterCodeRequired(ctx context.Context) (bool, error) {
	var setting model.Setting
	if err := s.db.WithContext(ctx).
		Where("config_key = ?", SettingRegisterCodeRequired).
		Take(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("读取系统设置失败: %w", err)
	}
	return setting.ConfigValue == "1", nil
}

// SetRegisterCodeRequired 更新开关，记录不存在时自动补建。
func (s *SettingService) SetRegisterCodeRequired(ctx context.Context, required bool) error {
	value := "0"
	if required {
		value = "1"
	}
	setting := model.Setting{ConfigKey: SettingRegisterCodeRequired, ConfigValue: "0"}
	if err := s.db.WithContext(ctx).
		Where("config_key = ?", SettingRegisterCodeRequired).
		FirstOrCreate(&setting).Error; err != nil {
		return fmt.Errorf("读取系统设置失败: %w", err)
	}
	setting.ConfigValue = value
	if err := s.db.WithContext(ctx).Save(&setting).Error; err != nil {
		return fmt.Errorf("保存系统设置失败: %w", err)
	}
	return nil
}
