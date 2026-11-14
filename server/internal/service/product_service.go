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

// DenominationInput 价格类别(面值)录入项。面值用字符串传入以适配 decimal 精度。
// sort 由录入顺序决定，无需前端传递。
type DenominationInput struct {
	Name      string `json:"name"`
	FaceValue string `json:"face_value"`
}

// CreateProductRequest 新增商品请求
type CreateProductRequest struct {
	Name          string              `json:"name" binding:"required,max=128"`
	Type          string              `json:"type" binding:"required,oneof=话费 卡券"`
	CoverImage    string              `json:"cover_image" binding:"required,max=512"`
	CostPrice     string              `json:"cost_price" binding:"required"`
	Discount      string              `json:"discount" binding:"required"`
	Denominations []DenominationInput `json:"denominations" binding:"required,min=1"`
	Description   string              `json:"description" binding:"max=512"`
	ChannelID     uint64              `json:"channel_id" binding:"required,min=1"`
	Status        int8                `json:"status" binding:"oneof=0 1"`
}

// UpdateProductRequest 更新商品请求（部分字段可选）
// Denominations 用指针区分「未传=不改」与「传空=清空」。
type UpdateProductRequest struct {
	Name          *string              `json:"name" binding:"omitempty,max=128"`
	Type          *string              `json:"type" binding:"omitempty,oneof=话费 卡券"`
	CoverImage    *string              `json:"cover_image" binding:"omitempty,max=512"`
	CostPrice     *string              `json:"cost_price"`
	Discount      *string              `json:"discount"`
	Denominations *[]DenominationInput `json:"denominations" binding:"omitempty,min=1"`
	Description   *string              `json:"description" binding:"omitempty,max=512"`
	ChannelID     *uint64              `json:"channel_id" binding:"omitempty,min=1"`
	Status        *int8                `json:"status" binding:"omitempty,oneof=0 1"`
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

// parseDiscountField 解析折扣(几折)：6 表示 6 折(付 60%，即 ×0.6)，取值范围 0 < d <= 10。
func parseDiscountField(s string) (model.Money, error) {
	d, err := parseMoneyField(s)
	if err != nil {
		return d, err
	}
	if !d.IsPositive() || d.GreaterThan(decimal.NewFromInt(10)) {
		return decimal.Decimal{}, fmt.Errorf("%w: 折扣必须在 0~10 之间(如 6 表示 6 折)", ErrInvalidPrice)
	}
	return d, nil
}

// buildDenominations 把录入项转成子表行，按录入顺序设置 sort，并校验面值大于 0。
func buildDenominations(productID uint64, inputs []DenominationInput) ([]model.ProductDenomination, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("%w: 至少需要一个价格类别", ErrInvalidPrice)
	}
	rows := make([]model.ProductDenomination, 0, len(inputs))
	for i, in := range inputs {
		faceValue, err := parseMoneyField(in.FaceValue)
		if err != nil {
			return nil, err
		}
		if !faceValue.IsPositive() {
			return nil, fmt.Errorf("%w: 面值必须大于 0", ErrInvalidPrice)
		}
		rows = append(rows, model.ProductDenomination{
			ProductID: productID,
			Name:      in.Name,
			FaceValue: faceValue,
			Sort:      i,
		})
	}
	return rows, nil
}

// orderDenominations Preload 时按 sort、id 升序排列价格类别。
func orderDenominations(db *gorm.DB) *gorm.DB {
	return db.Order("sort ASC, id ASC")
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

// Create 新增商品。上架（status=1）前校验绑定渠道有效性；价格类别在事务内一并写入。
func (s *ProductService) Create(ctx context.Context, req *CreateProductRequest) (*model.Product, error) {
	costPrice, err := parseMoneyField(req.CostPrice)
	if err != nil {
		return nil, err
	}
	discount, err := parseDiscountField(req.Discount)
	if err != nil {
		return nil, err
	}
	if req.Status == model.ProductStatusOnline {
		if err := s.ensureChannelUsable(ctx, req.ChannelID); err != nil {
			return nil, err
		}
	}
	p := &model.Product{
		Name:        req.Name,
		Type:        req.Type,
		CoverImage:  req.CoverImage,
		CostPrice:   costPrice,
		Discount:    discount,
		Description: req.Description,
		ChannelID:   req.ChannelID,
		Status:      req.Status,
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return fmt.Errorf("创建商品失败: %w", err)
		}
		rows, err := buildDenominations(p.ID, req.Denominations)
		if err != nil {
			return err
		}
		if err := tx.Create(&rows).Error; err != nil {
			return fmt.Errorf("创建价格类别失败: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Preload("Denominations", orderDenominations).First(p, p.ID).Error; err != nil {
		return nil, fmt.Errorf("刷新商品数据失败: %w", err)
	}
	return p, nil
}

// List 分页查询商品，支持按名称模糊匹配、状态精确过滤；价格类别随商品一起 Preload。
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
		Preload("Denominations", orderDenominations).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询商品列表失败: %w", err)
	}
	return list, total, nil
}

// Update 部分更新商品。若最终状态为上架，校验绑定渠道有效性；
// 价格类别仅在请求显式提供时整体替换（先硬删旧行再按新顺序插入）。
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
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ChannelID != nil {
		updates["channel_id"] = *req.ChannelID
	}
	if req.Discount != nil {
		v, err := parseDiscountField(*req.Discount)
		if err != nil {
			return nil, err
		}
		updates["discount"] = v
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

	// 价格类别：仅在显式传入时构建并替换
	var newDenominations []model.ProductDenomination
	if req.Denominations != nil {
		rows, err := buildDenominations(id, *req.Denominations)
		if err != nil {
			return nil, err
		}
		newDenominations = rows
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&p).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新商品失败: %w", err)
			}
		}
		if req.Denominations != nil {
			if err := tx.Unscoped().Where("product_id = ?", id).Delete(&model.ProductDenomination{}).Error; err != nil {
				return fmt.Errorf("清除旧价格类别失败: %w", err)
			}
			if len(newDenominations) > 0 {
				if err := tx.Create(&newDenominations).Error; err != nil {
					return fmt.Errorf("保存价格类别失败: %w", err)
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 重新读取以保证返回最新数据（含价格类别）
	if err := s.db.WithContext(ctx).Preload("Denominations", orderDenominations).First(&p, id).Error; err != nil {
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
