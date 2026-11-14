package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"niulai/internal/model"
)

// 商品相关业务错误
var (
	ErrProductNotFound = errors.New("商品不存在")
	// ErrChannelUnusable 上架时绑定的渠道无效（不存在或处于熔断状态）
	ErrChannelUnusable = errors.New("商品上架要求绑定有效（可用状态）的上游渠道")
	// ErrInvalidPrice 价格格式非法或为负
	ErrInvalidPrice = errors.New("价格格式非法")
)

// ProductService 商品管理业务逻辑
type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

// CreateProductRequest 新增商品请求
type CreateProductRequest struct {
	Name          string `json:"name" binding:"required,max=128"`
	Type          string `json:"type" binding:"required,oneof=话费 卡券"`
	CoverImage    string `json:"cover_image" binding:"required,max=512"`
	SalePrice     string `json:"sale_price" binding:"required"`
	CostPrice     string `json:"cost_price" binding:"required"`
	DiscountLabel string `json:"discount_label" binding:"max=32"`
	Description   string `json:"description" binding:"max=512"`
	ChannelID     uint64 `json:"channel_id" binding:"required,min=1"`
	Status        int8   `json:"status" binding:"oneof=0 1"`
}

// UpdateProductRequest 更新商品请求（部分字段可选）
type UpdateProductRequest struct {
	Name          *string `json:"name" binding:"omitempty,max=128"`
	Type          *string `json:"type" binding:"omitempty,oneof=话费 卡券"`
	CoverImage    *string `json:"cover_image" binding:"omitempty,max=512"`
	SalePrice     *string `json:"sale_price"`
	CostPrice     *string `json:"cost_price"`
	DiscountLabel *string `json:"discount_label" binding:"omitempty,max=32"`
	Description   *string `json:"description" binding:"omitempty,max=512"`
	ChannelID     *uint64 `json:"channel_id" binding:"omitempty,min=1"`
	Status        *int8   `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateProductStatusRequest 上下架切换请求
type UpdateProductStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

// ListProductRequest 商品分页查询
type ListProductRequest struct {
	Name     string `form:"name"`
	Status   *int8  `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// parseMoneyField 将字符串金额解析为 decimal，空串视为 0，负数视为非法。
func parseMoneyField(s string) (model.Money, error) {
	if s == "" {
		return decimal.Zero, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("%w: %s", ErrInvalidPrice, err.Error())
	}
	if d.IsNegative() {
		return decimal.Decimal{}, fmt.Errorf("%w: 金额不能为负数", ErrInvalidPrice)
	}
	return d, nil
}

// ensureChannelUsable 上架前校验：绑定渠道必须存在且处于可用（非熔断）状态。
func (s *ProductService) ensureChannelUsable(ctx context.Context, channelID uint64) error {
	if channelID == 0 {
		return ErrChannelUnusable
	}
	var ch model.Channel
	if err := s.db.WithContext(ctx).First(&ch, channelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelUnusable
		}
		return fmt.Errorf("查询渠道失败: %w", err)
	}
	if ch.Status != model.ChannelStatusOnline {
		return ErrChannelUnusable
	}
	return nil
}

// Create 新增商品。上架（status=1）前校验绑定渠道有效性。
func (s *ProductService) Create(ctx context.Context, req *CreateProductRequest) (*model.Product, error) {
	salePrice, err := parseMoneyField(req.SalePrice)
	if err != nil {
		return nil, err
	}
	costPrice, err := parseMoneyField(req.CostPrice)
	if err != nil {
		return nil, err
	}
	if req.Status == model.ProductStatusOnline {
		if err := s.ensureChannelUsable(ctx, req.ChannelID); err != nil {
			return nil, err
		}
	}
	p := &model.Product{
		Name:          req.Name,
		Type:          req.Type,
		CoverImage:    req.CoverImage,
		SalePrice:     salePrice,
		CostPrice:     costPrice,
		DiscountLabel: req.DiscountLabel,
		Description:   req.Description,
		ChannelID:     req.ChannelID,
		Status:        req.Status,
	}
	if err := s.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, fmt.Errorf("创建商品失败: %w", err)
	}
	return p, nil
}

// List 分页查询商品，支持按名称模糊匹配、状态精确过滤。
func (s *ProductService) List(ctx context.Context, req *ListProductRequest) ([]model.Product, int64, error) {
	var (
		list  []model.Product
		total int64
	)
	q := s.db.WithContext(ctx).Model(&model.Product{})
	if req.Name != "" {
		q = q.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计商品数量失败: %w", err)
	}
	if err := q.Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询商品列表失败: %w", err)
	}
	return list, total, nil
}

// Update 部分更新商品。若最终状态为上架，校验绑定渠道有效性。
func (s *ProductService) Update(ctx context.Context, id uint64, req *UpdateProductRequest) (*model.Product, error) {
	var p model.Product
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("查询商品失败: %w", err)
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.CoverImage != nil {
		updates["cover_image"] = *req.CoverImage
	}
	if req.DiscountLabel != nil {
		updates["discount_label"] = *req.DiscountLabel
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ChannelID != nil {
		updates["channel_id"] = *req.ChannelID
	}
	if req.SalePrice != nil {
		v, err := parseMoneyField(*req.SalePrice)
		if err != nil {
			return nil, err
		}
		updates["sale_price"] = v
	}
	if req.CostPrice != nil {
		v, err := parseMoneyField(*req.CostPrice)
		if err != nil {
			return nil, err
		}
		updates["cost_price"] = v
	}

	// 计算更新后的最终状态与渠道，用于上架安全校验
	newStatus := p.Status
	if req.Status != nil {
		newStatus = *req.Status
		updates["status"] = *req.Status
	}
	newChannelID := p.ChannelID
	if req.ChannelID != nil {
		newChannelID = *req.ChannelID
	}
	if newStatus == model.ProductStatusOnline {
		if err := s.ensureChannelUsable(ctx, newChannelID); err != nil {
			return nil, err
		}
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&p).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新商品失败: %w", err)
		}
	}
	// 重新读取以保证返回最新数据（含未变更的字段）
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, fmt.Errorf("刷新商品数据失败: %w", err)
	}
	return &p, nil
}

// UpdateStatus 上下架切换。上架前校验绑定渠道有效性。
func (s *ProductService) UpdateStatus(ctx context.Context, id uint64, status int8) (*model.Product, error) {
	var p model.Product
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("查询商品失败: %w", err)
	}
	if status == model.ProductStatusOnline {
		if err := s.ensureChannelUsable(ctx, p.ChannelID); err != nil {
			return nil, err
		}
	}
	if err := s.db.WithContext(ctx).Model(&p).Update("status", status).Error; err != nil {
		return nil, fmt.Errorf("更新商品状态失败: %w", err)
	}
	p.Status = status
	return &p, nil
}
