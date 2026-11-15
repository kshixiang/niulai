package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"niulai/internal/model"
)

// WalletView 当前用户资金账户对外视图。金额以两位小数字符串返回，
// 避免前端用浮点累加造成精度误差。
type WalletView struct {
	Balance       string `json:"balance"`
	FrozenBalance string `json:"frozen_balance"`
}

// WalletService C 端资金账户查询服务。
type WalletService struct {
	db *gorm.DB
}

func NewWalletService(db *gorm.DB) *WalletService {
	return &WalletService{db: db}
}

// GetByUser 按用户ID查询资金账户。尚未开通账户（无记录）时按零余额返回，
// 读接口不写库，账户的创建由后续资金业务统一负责。
func (s *WalletService) GetByUser(ctx context.Context, userID uint64) (*WalletView, error) {
	var w model.Wallet
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&w).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &WalletView{Balance: "0.00", FrozenBalance: "0.00"}, nil
		}
		return nil, fmt.Errorf("查询资金账户失败: %w", err)
	}
	return &WalletView{
		Balance:       w.Balance.StringFixed(2),
		FrozenBalance: w.FrozenBalance.StringFixed(2),
	}, nil
}
