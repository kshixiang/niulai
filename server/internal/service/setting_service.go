package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"niulai/internal/model"
)

// 系统配置键
const (
	// SettingRegisterCodeRequired 注册码是否强制填写的配置键。
	SettingRegisterCodeRequired = "register_code_required"
	// SettingUSDTCNYRate USDT→CNY 汇率：1 USDT 可兑换的 CNY(RMB) 数量。
	// 订单金额按 CNY 结算、充值余额按 USDT 结算，下单时按此汇率换算后校验余额。
	SettingUSDTCNYRate = "usdt_cny_rate"
	// DefaultUSDTCNYRate 汇率缺省值（配置缺失或非法时回退，保证下单链路可用）。
	DefaultUSDTCNYRate = "6.69"
)

// ErrInvalidExchangeRate 汇率取值不合法。
var ErrInvalidExchangeRate = errors.New("汇率须为大于0且不超过1000、最多2位小数的数值")

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

// GetUSDTCNYRate 读取当前汇率。记录缺失或值非法时回退到默认值，
// 避免运营尚未配置或误填时下单链路不可用。
func (s *SettingService) GetUSDTCNYRate(ctx context.Context) (model.Money, error) {
	var setting model.Setting
	err := s.db.WithContext(ctx).
		Where("config_key = ?", SettingUSDTCNYRate).
		Take(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.NewMoney(DefaultUSDTCNYRate), nil
		}
		return model.Money{}, fmt.Errorf("读取系统设置失败: %w", err)
	}
	rate, err := decimal.NewFromString(strings.TrimSpace(setting.ConfigValue))
	if err != nil || !rate.IsPositive() {
		return model.NewMoney(DefaultUSDTCNYRate), nil
	}
	return rate.Round(2), nil
}

// SetUSDTCNYRate 更新汇率。取值须为正数、不超过 1000、最多 2 位小数；
// 统一按 StringFixed(2) 落库，保证展示与换算使用同一精度。
func (s *SettingService) SetUSDTCNYRate(ctx context.Context, raw string) error {
	rate, err := decimal.NewFromString(strings.TrimSpace(raw))
	if err != nil || !rate.IsPositive() ||
		rate.GreaterThan(decimal.NewFromInt(1000)) || !rate.Equal(rate.Round(2)) {
		return ErrInvalidExchangeRate
	}
	value := rate.StringFixed(2)
	setting := model.Setting{ConfigKey: SettingUSDTCNYRate, ConfigValue: value}
	if err := s.db.WithContext(ctx).
		Where("config_key = ?", SettingUSDTCNYRate).
		FirstOrCreate(&setting).Error; err != nil {
		return fmt.Errorf("读取系统设置失败: %w", err)
	}
	setting.ConfigValue = value
	if err := s.db.WithContext(ctx).Save(&setting).Error; err != nil {
		return fmt.Errorf("保存系统设置失败: %w", err)
	}
	return nil
}
