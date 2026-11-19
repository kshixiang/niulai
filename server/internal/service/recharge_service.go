package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"niulai/internal/model"
)

const SettingRechargeAddress = "recharge_trc20_address"

var (
	ErrInvalidRecharge        = errors.New("充值参数不正确")
	ErrRechargeUnavailable    = errors.New("充值通道暂未开放，请稍后重试")
	ErrRechargeAddressChanged = errors.New("收款地址已更新，请刷新页面并重新核对转账地址")
	ErrRechargeDuplicate      = errors.New("该交易已申报，请勿重复提交；如有疑问请联系客服")
	ErrRechargeReviewed       = errors.New("该充值已审核，请刷新记录，勿重复操作")
	rechargeAmountPattern     = regexp.MustCompile(`^\d{1,8}(\.\d{1,6})?$`)
	rechargeHashPattern       = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
)

func parseRechargeAmount(raw string) (model.Money, error) {
	raw = strings.TrimSpace(raw)
	if !rechargeAmountPattern.MatchString(raw) {
		return model.Money{}, fmt.Errorf("%w：金额最多8位整数、6位小数", ErrInvalidRecharge)
	}
	amount, err := decimal.NewFromString(raw)
	if err != nil || amount.LessThan(decimal.NewFromInt(1)) || amount.GreaterThan(model.NewMoney("99999999.99")) {
		return model.Money{}, fmt.Errorf("%w：单笔金额须在1至99999999.99 USDT之间", ErrInvalidRecharge)
	}
	return amount.Truncate(2), nil
}

func normalizeRechargeHash(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if !rechargeHashPattern.MatchString(raw) {
		return "", fmt.Errorf("%w：请输入64位十六进制交易哈希", ErrInvalidRecharge)
	}
	return strings.ToLower(raw), nil
}

// validateRechargeAddress 解码 Base58Check，检查 TRON 主网前缀与双 SHA256 校验和。
func validateRechargeAddress(raw string) error {
	invalid := fmt.Errorf("%w：请输入有效的TRON主网收款地址", ErrInvalidRecharge)
	if len(raw) != 34 || raw[0] != 'T' {
		return invalid
	}
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	n := new(big.Int)
	for _, c := range raw {
		idx := strings.IndexRune(alphabet, c)
		if idx < 0 {
			return invalid
		}
		n.Mul(n, big.NewInt(58))
		n.Add(n, big.NewInt(int64(idx)))
	}
	decoded := n.Bytes()
	if len(decoded) != 25 || decoded[0] != 0x41 {
		return invalid
	}
	h := sha256.Sum256(decoded[:21])
	h = sha256.Sum256(h[:])
	if !bytes.Equal(h[:4], decoded[21:]) {
		return invalid
	}
	return nil
}

func (s *SettingService) GetRechargeAddress(ctx context.Context) (string, error) {
	var row model.Setting
	err := s.db.WithContext(ctx).Where("config_key = ?", SettingRechargeAddress).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("读取充值地址失败: %w", err)
	}
	return strings.TrimSpace(row.ConfigValue), nil
}

func (s *SettingService) SetRechargeAddress(ctx context.Context, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		if err := validateRechargeAddress(raw); err != nil {
			return err
		}
	}
	row := model.Setting{ConfigKey: SettingRechargeAddress, ConfigValue: raw}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"config_value": raw}),
	}).Create(&row).Error
}

type RechargeConfig struct {
	Address   string `json:"address"`
	Network   string `json:"network"`
	Currency  string `json:"currency"`
	MinAmount string `json:"min_amount"`
	Enabled   bool   `json:"enabled"`
}

type CreateRechargeRequest struct {
	Address string `json:"address" binding:"required,max=34"`
	Amount  string `json:"amount" binding:"required,max=32"`
	TxHash  string `json:"tx_hash" binding:"required,len=64"`
}

type ListRechargeRequest struct {
	Page     int   `form:"page"`
	PageSize int   `form:"page_size"`
	Status   *int8 `form:"status" binding:"omitempty,min=0,max=2"`
}

type ReviewRechargeRequest struct {
	Status int8   `json:"status" binding:"required,oneof=1 2"`
	Amount string `json:"amount" binding:"max=32"`
	Remark string `json:"remark" binding:"max=255"`
}

type RechargeService struct {
	db       *gorm.DB
	settings *SettingService
}

func NewRechargeService(db *gorm.DB, settings *SettingService) *RechargeService {
	return &RechargeService{db: db, settings: settings}
}

func (s *RechargeService) Config(ctx context.Context) (*RechargeConfig, error) {
	address, err := s.settings.GetRechargeAddress(ctx)
	if err != nil {
		return nil, err
	}
	enabled := address != "" && validateRechargeAddress(address) == nil
	if !enabled {
		address = ""
	}
	return &RechargeConfig{Address: address, Network: "TRC-20", Currency: "USDT", MinAmount: "1", Enabled: enabled}, nil
}

func (s *RechargeService) Create(ctx context.Context, userID uint64, req *CreateRechargeRequest) (*model.Recharge, error) {
	amount, err := parseRechargeAmount(req.Amount)
	if err != nil {
		return nil, err
	}
	hash, err := normalizeRechargeHash(req.TxHash)
	if err != nil {
		return nil, err
	}
	cfg, err := s.Config(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, ErrRechargeUnavailable
	}
	if req.Address != cfg.Address {
		return nil, ErrRechargeAddressChanged
	}
	// 不信任客户端身份：userID 只来自 JWT，并重新检查账号是否可用。
	var user model.User
	if err := s.db.WithContext(ctx).Select("id", "status").First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.Status != 1 {
		return nil, ErrUserDisabled
	}
	row := model.Recharge{UserID: userID, Address: cfg.Address, TxHash: hash, ReportedAmount: strings.TrimSpace(req.Amount), Amount: amount, Status: model.RechargePending}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		var myErr *mysqlDriver.MySQLError
		if errors.As(err, &myErr) && myErr.Number == 1062 {
			return nil, ErrRechargeDuplicate
		}
		return nil, fmt.Errorf("保存充值记录失败: %w", err)
	}
	return &row, nil
}

// List userID 非0时始终过滤本人；0只允许由后台控制器使用。
func (s *RechargeService) List(ctx context.Context, userID uint64, req *ListRechargeRequest) ([]model.Recharge, int64, error) {
	q := s.db.WithContext(ctx).Model(&model.Recharge{})
	if userID != 0 {
		q = q.Where("user_id = ?", userID)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]model.Recharge, 0)
	err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error
	return rows, total, err
}

// Review 锁定充值记录后再锁钱包；状态、余额、流水原子提交。
func (s *RechargeService) Review(ctx context.Context, id uint64, req *ReviewRechargeRequest) (*model.Recharge, error) {
	if req.Status != model.RechargeApproved && req.Status != model.RechargeRejected {
		return nil, ErrInvalidRecharge
	}
	remark := strings.TrimSpace(req.Remark)
	if utf8.RuneCountInString(remark) > 255 {
		return nil, ErrInvalidRecharge
	}
	if req.Status == model.RechargeRejected && remark == "" {
		return nil, fmt.Errorf("%w：请填写驳回原因", ErrInvalidRecharge)
	}
	if req.Status == model.RechargeApproved && remark == "" {
		return nil, fmt.Errorf("%w：请记录付款归属核验说明，公开交易哈希不能作为归属证明", ErrInvalidRecharge)
	}
	var amount model.Money
	if req.Status == model.RechargeApproved {
		var err error
		amount, err = parseRechargeAmount(req.Amount)
		if err != nil {
			return nil, err
		}
	}
	var row model.Recharge
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
			return err
		}
		if row.Status != model.RechargePending {
			return ErrRechargeReviewed
		}
		if req.Status == model.RechargeApproved {
			// 新账号允许首次充值时创建零余额钱包。唯一 user_id 防止并发创建。
			wallet := model.Wallet{UserID: row.UserID, Balance: decimal.Zero, FrozenBalance: decimal.Zero, Version: 1}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&wallet).Error; err != nil {
				return err
			}
			// GORM 会把 Create 返回的主键带入后续 First，清空后仅按 user_id 查询。
			wallet = model.Wallet{}
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", row.UserID).First(&wallet).Error; err != nil {
				return err
			}
			after := wallet.Balance.Add(amount)
			if after.GreaterThan(model.NewMoney("99999999.99")) {
				return fmt.Errorf("%w：入账后余额超出账户上限", ErrInvalidRecharge)
			}
			if err := tx.Model(&model.Wallet{}).Where("user_id = ?", row.UserID).Updates(map[string]interface{}{"balance": after, "version": gorm.Expr("version + 1")}).Error; err != nil {
				return err
			}
			ledger := model.WalletTransaction{UserID: row.UserID, Type: model.WalletTxRecharge, Amount: amount, BalanceBefore: wallet.Balance, BalanceAfter: after, FrozenBefore: wallet.FrozenBalance, FrozenAfter: wallet.FrozenBalance, Ref: fmt.Sprintf("RECHARGE-%d", row.ID), Remark: "USDT/TRC-20充值到账"}
			if err := tx.Create(&ledger).Error; err != nil {
				return err
			}
			row.Amount = amount
		}
		now := time.Now()
		row.Status, row.Remark, row.ReviewedAt = req.Status, remark, &now
		return tx.Save(&row).Error
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}
