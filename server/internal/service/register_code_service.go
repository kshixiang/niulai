package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"niulai/internal/model"
)

// 注册码状态
const (
	RegisterCodeEnabled  int8 = 1
	RegisterCodeDisabled int8 = 0
	maxGenerateCount          = 100
)

var (
	// ErrInviterNotFound 注册码绑定的邀请人不存在或已被封禁。
	ErrInviterNotFound = errors.New("注册码对应的邀请人不存在")
	// ErrRegisterCodeNotFound 更新状态时找不到对应注册码。
	ErrRegisterCodeNotFound = errors.New("注册码不存在")
)

// RegisterCodeService 注册码生成与校验。
type RegisterCodeService struct {
	db *gorm.DB
}

func NewRegisterCodeService(db *gorm.DB) *RegisterCodeService {
	return &RegisterCodeService{db: db}
}

// Generate 批量生成注册码。inviterID 非 0 时先校验邀请人存在。
func (s *RegisterCodeService) Generate(ctx context.Context, count int, inviterID uint64) ([]model.RegisterCode, error) {
	if count < 1 || count > maxGenerateCount {
		return nil, fmt.Errorf("生成数量需在 1-%d 之间", maxGenerateCount)
	}
	if inviterID != 0 && !s.inviterExists(ctx, inviterID) {
		return nil, ErrInviterNotFound
	}

	codes := make([]model.RegisterCode, 0, count)
	for i := 0; i < count; i++ {
		code, err := randomCode()
		if err != nil {
			return nil, err
		}
		codes = append(codes, model.RegisterCode{
			Code:      code,
			InviterID: inviterID,
			Status:    RegisterCodeEnabled,
		})
	}
	if err := s.db.WithContext(ctx).Create(&codes).Error; err != nil {
		return nil, fmt.Errorf("生成注册码失败: %w", err)
	}
	return codes, nil
}

// List 分页查询注册码。
func (s *RegisterCodeService) List(ctx context.Context, page, pageSize int) ([]model.RegisterCode, int64, error) {
	var (
		list  []model.RegisterCode
		total int64
	)
	if err := s.db.WithContext(ctx).Model(&model.RegisterCode{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计注册码失败: %w", err)
	}
	if err := s.db.WithContext(ctx).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询注册码失败: %w", err)
	}
	return list, total, nil
}

// UpdateStatus 启用或禁用注册码。
func (s *RegisterCodeService) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	res := s.db.WithContext(ctx).Model(&model.RegisterCode{}).
		Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("更新注册码状态失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrRegisterCodeNotFound
	}
	return nil
}

// Validate 注册时校验：码存在且可用；绑定了邀请人时该用户必须存在且状态正常。
func (s *RegisterCodeService) Validate(ctx context.Context, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return ErrRegisterCodeInvalid
	}
	var rc model.RegisterCode
	if err := s.db.WithContext(ctx).Where("code = ?", code).Take(&rc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRegisterCodeInvalid
		}
		return fmt.Errorf("查询注册码失败: %w", err)
	}
	if rc.Status != RegisterCodeEnabled {
		return ErrRegisterCodeInvalid
	}
	if rc.InviterID != 0 && !s.inviterExists(ctx, rc.InviterID) {
		return ErrInviterNotFound
	}
	return nil
}

// inviterExists 校验邀请人：存在、正常状态、未软删除。
func (s *RegisterCodeService) inviterExists(ctx context.Context, userID uint64) bool {
	var count int64
	return s.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ? AND status = ?", userID, model.UserStatusNormal).
		Count(&count).Error == nil && count > 0
}

func randomCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成随机注册码失败: %w", err)
	}
	return "RC" + strings.ToUpper(hex.EncodeToString(b)), nil
}
