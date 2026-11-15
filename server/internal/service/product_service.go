package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

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
	// ErrInvalidProductField 下单填写项配置非法
	ErrInvalidProductField = errors.New("下单填写项配置非法")
)

// 下单填写项类型白名单
const (
	OrderFieldTypeText   = "text"
	OrderFieldTypeNumber = "number"
	OrderFieldTypeSelect = "select"
)

var fieldKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// carrierOptions 运营商下拉选项（JSON 数组字符串，与 product_order_fields.options 列格式一致）。
const carrierOptions = `["移动","电信","联通","广电"]`

// defaultPhoneFields 话费商品必须填写的字段。后台若已配置同名 field_key，则以后台配置为准。
var defaultPhoneFields = []model.ProductOrderField{
	{FieldKey: "phone", Label: "手机号", Type: OrderFieldTypeText, Required: 1, Placeholder: "请输入充值手机号"},
	{FieldKey: "username", Label: "用户名", Type: OrderFieldTypeText, Required: 1, Placeholder: "请输入用户名"},
	{FieldKey: "carrier", Label: "运营商", Type: OrderFieldTypeSelect, Required: 1, Placeholder: "请选择运营商", Options: carrierOptions},
}

// appendPhoneFields 为话费商品补齐后台未配置的必填项，并排在最前。
// 商品详情（渲染下单页）与下单校验共用，保证两端字段与必填规则一致。
func appendPhoneFields(fields []model.ProductOrderField, productType string) []model.ProductOrderField {
	if productType != model.ProductTypePhone {
		return fields
	}
	configured := make(map[string]bool, len(fields))
	for _, f := range fields {
		configured[f.FieldKey] = true
	}
	missing := make([]model.ProductOrderField, 0, len(defaultPhoneFields))
	for _, def := range defaultPhoneFields {
		if !configured[def.FieldKey] {
			missing = append(missing, def)
		}
	}
	if len(missing) > 0 {
		fields = append(missing, fields...)
	}
	return fields
}

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

// ProductFieldInput 下单填写项录入项。
type ProductFieldInput struct {
	FieldKey    string   `json:"field_key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options"`
}

// CreateProductRequest 新增商品请求
type CreateProductRequest struct {
	Name            string              `json:"name" binding:"required,max=128"`
	Type            string              `json:"type" binding:"required,oneof=话费 卡券"`
	CoverImage      string              `json:"cover_image" binding:"required,max=512"`
	ChannelDiscount string              `json:"channel_discount" binding:"required"`
	Discount        string              `json:"discount" binding:"required"`
	Denominations   []DenominationInput `json:"denominations" binding:"required,min=1"`
	Fields          []ProductFieldInput `json:"fields" binding:"omitempty"`
	Description     string              `json:"description" binding:"max=512"`
	ChannelID       uint64              `json:"channel_id" binding:"required,min=1"`
	Status          int8                `json:"status" binding:"oneof=0 1"`
}

// UpdateProductRequest 更新商品请求（部分字段可选）
// Denominations / Fields 用指针区分「未传=不改」与「传空=清空」。
type UpdateProductRequest struct {
	Name            *string              `json:"name" binding:"omitempty,max=128"`
	Type            *string              `json:"type" binding:"omitempty,oneof=话费 卡券"`
	CoverImage      *string              `json:"cover_image" binding:"omitempty,max=512"`
	ChannelDiscount *string              `json:"channel_discount"`
	Discount        *string              `json:"discount"`
	Denominations   *[]DenominationInput `json:"denominations" binding:"omitempty,min=1"`
	Fields          *[]ProductFieldInput `json:"fields" binding:"omitempty"`
	Description     *string              `json:"description" binding:"omitempty,max=512"`
	ChannelID       *uint64              `json:"channel_id" binding:"omitempty,min=1"`
	Status          *int8                `json:"status" binding:"omitempty,oneof=0 1"`
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

// ---------- 后台响应视图（含内部经营数据，仅 admin 接口使用） ----------

// AdminDenominationView 后台价格类别视图，含渠道价/售价/盈利（decimal 服务端计算）。
type AdminDenominationView struct {
	ID           uint64      `json:"id"`
	Name         string      `json:"name"`
	FaceValue    model.Money `json:"face_value"`
	ChannelPrice model.Money `json:"channel_price"`
	SellPrice    model.Money `json:"sell_price"`
	Profit       model.Money `json:"profit"`
	Sort         int         `json:"sort"`
}

// AdminFieldView 后台下单填写项视图，options 已解析为数组。
type AdminFieldView struct {
	ID          uint64   `json:"id"`
	FieldKey    string   `json:"field_key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    int8     `json:"required"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options"`
	Sort        int      `json:"sort"`
}

// AdminProductView 后台商品视图（含渠道折扣与盈利，禁止用于公开接口）。
type AdminProductView struct {
	ID              uint64                  `json:"id"`
	Name            string                  `json:"name"`
	Type            string                  `json:"type"`
	CoverImage      string                  `json:"cover_image"`
	ChannelDiscount model.Money             `json:"channel_discount"`
	Discount        model.Money             `json:"discount"`
	Description     string                  `json:"description"`
	ChannelID       uint64                  `json:"channel_id"`
	Status          int8                    `json:"status"`
	Denominations   []AdminDenominationView `json:"denominations"`
	Fields          []AdminFieldView        `json:"fields"`
}

// ---------- 工具函数 ----------

// faceToPrice 售价/渠道价统一计算：面值 × 折扣 ÷ 10，decimal.Round(2)；异常折扣退化为不打折。
func faceToPrice(face, discount model.Money) model.Money {
	ten := decimal.NewFromInt(10)
	if !discount.IsPositive() {
		discount = ten
	}
	return face.Mul(discount).Div(ten).Round(2)
}

// parseOptions 把 options JSON 数组字符串解析为 []string，非法或空返回空数组。
func parseOptions(s string) []string {
	if s == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil || out == nil {
		return []string{}
	}
	return out
}

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

// buildOrderFields 把录入项转成下单填写项子表行，校验标识/类型/选项并序列化 options。
func buildOrderFields(productID uint64, inputs []ProductFieldInput) ([]model.ProductOrderField, error) {
	rows := make([]model.ProductOrderField, 0, len(inputs))
	seen := make(map[string]struct{}, len(inputs))
	for i, in := range inputs {
		key := strings.TrimSpace(in.FieldKey)
		if !fieldKeyPattern.MatchString(key) {
			return nil, fmt.Errorf("%w: 字段标识须以小写字母开头，仅含小写字母/数字/下划线(如 phone)", ErrInvalidProductField)
		}
		if _, dup := seen[key]; dup {
			return nil, fmt.Errorf("%w: 字段标识 %s 重复", ErrInvalidProductField, key)
		}
		seen[key] = struct{}{}
		label := strings.TrimSpace(in.Label)
		if label == "" || len([]rune(label)) > 128 {
			return nil, fmt.Errorf("%w: 中文名必填且不超过 128 字", ErrInvalidProductField)
		}
		switch in.Type {
		case OrderFieldTypeText, OrderFieldTypeNumber, OrderFieldTypeSelect:
		default:
			return nil, fmt.Errorf("%w: 字段类型仅支持 text/number/select", ErrInvalidProductField)
		}
		opts := in.Options
		if in.Type == OrderFieldTypeSelect {
			cleaned := make([]string, 0, len(opts))
			for _, o := range opts {
				if v := strings.TrimSpace(o); v != "" {
					cleaned = append(cleaned, v)
				}
			}
			if len(cleaned) == 0 {
				return nil, fmt.Errorf("%w: 下拉类型必须至少配置一个选项", ErrInvalidProductField)
			}
			opts = cleaned
		} else {
			opts = []string{}
		}
		optJSON, err := json.Marshal(opts)
		if err != nil {
			return nil, fmt.Errorf("%w: 选项序列化失败", ErrInvalidProductField)
		}
		placeholder := in.Placeholder
		if len([]rune(placeholder)) > 255 {
			return nil, fmt.Errorf("%w: 输入提示语不超过 255 字", ErrInvalidProductField)
		}
		required := int8(0)
		if in.Required {
			required = 1
		}
		rows = append(rows, model.ProductOrderField{
			ProductID:   productID,
			FieldKey:    key,
			Label:       label,
			Type:        in.Type,
			Required:    required,
			Placeholder: placeholder,
			Options:     string(optJSON),
			Sort:        i,
		})
	}
	return rows, nil
}

func orderDenominations(db *gorm.DB) *gorm.DB {
	return db.Order("sort ASC, id ASC")
}

func orderFields(db *gorm.DB) *gorm.DB {
	return db.Order("sort ASC, id ASC")
}

// toAdminView 把 Product(含 Preload 关联) 映射为后台视图，并计算渠道价/售价/盈利。
func toAdminView(p model.Product) AdminProductView {
	v := AdminProductView{
		ID:              p.ID,
		Name:            p.Name,
		Type:            p.Type,
		CoverImage:      p.CoverImage,
		ChannelDiscount: p.ChannelDiscount,
		Discount:        p.Discount,
		Description:     p.Description,
		ChannelID:       p.ChannelID,
		Status:          p.Status,
		Denominations:   make([]AdminDenominationView, 0, len(p.Denominations)),
		Fields:          make([]AdminFieldView, 0, len(p.Fields)),
	}
	for _, d := range p.Denominations {
		channelPrice := faceToPrice(d.FaceValue, p.ChannelDiscount)
		sellPrice := faceToPrice(d.FaceValue, p.Discount)
		v.Denominations = append(v.Denominations, AdminDenominationView{
			ID:           d.ID,
			Name:         d.Name,
			FaceValue:    d.FaceValue,
			ChannelPrice: channelPrice,
			SellPrice:    sellPrice,
			Profit:       sellPrice.Sub(channelPrice),
			Sort:         d.Sort,
		})
	}
	for _, f := range p.Fields {
		v.Fields = append(v.Fields, AdminFieldView{
			ID:          f.ID,
			FieldKey:    f.FieldKey,
			Label:       f.Label,
			Type:        f.Type,
			Required:    f.Required,
			Placeholder: f.Placeholder,
			Options:     parseOptions(f.Options),
			Sort:        f.Sort,
		})
	}
	return v
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

// Create 新增商品。上架（status=1）前校验绑定渠道有效性；价格类别与填写项在事务内一并写入。
func (s *ProductService) Create(ctx context.Context, req *CreateProductRequest) (*model.Product, error) {
	channelDiscount, err := parseDiscountField(req.ChannelDiscount)
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
	fieldRows, err := buildOrderFields(0, req.Fields)
	if err != nil {
		return nil, err
	}
	p := &model.Product{
		Name:            req.Name,
		Type:            req.Type,
		CoverImage:      req.CoverImage,
		ChannelDiscount: channelDiscount,
		Discount:        discount,
		Description:     req.Description,
		ChannelID:       req.ChannelID,
		Status:          req.Status,
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
		if len(fieldRows) > 0 {
			for i := range fieldRows {
				fieldRows[i].ProductID = p.ID
			}
			if err := tx.Create(&fieldRows).Error; err != nil {
				return fmt.Errorf("创建下单填写项失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := s.reloadProduct(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// List 分页查询商品，支持按名称模糊匹配、状态精确过滤；返回含盈利的后台视图。
func (s *ProductService) List(ctx context.Context, req *ListProductRequest) ([]AdminProductView, int64, error) {
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
		Preload("Fields", orderFields).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询商品列表失败: %w", err)
	}
	views := make([]AdminProductView, 0, len(list))
	for _, p := range list {
		views = append(views, toAdminView(p))
	}
	return views, total, nil
}

// Update 部分更新商品。若最终状态为上架，校验绑定渠道有效性；
// 价格类别与填写项仅在请求显式提供时整体替换（硬删旧行再按新顺序插入）。
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
	if req.ChannelDiscount != nil {
		v, err := parseDiscountField(*req.ChannelDiscount)
		if err != nil {
			return nil, err
		}
		updates["channel_discount"] = v
	}
	if req.Discount != nil {
		v, err := parseDiscountField(*req.Discount)
		if err != nil {
			return nil, err
		}
		updates["discount"] = v
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
	// 下单填写项：仅在显式传入时构建并替换（传空切片=清空）
	var newFields []model.ProductOrderField
	if req.Fields != nil {
		rows, err := buildOrderFields(id, *req.Fields)
		if err != nil {
			return nil, err
		}
		newFields = rows
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
		if req.Fields != nil {
			if err := tx.Unscoped().Where("product_id = ?", id).Delete(&model.ProductOrderField{}).Error; err != nil {
				return fmt.Errorf("清除旧下单填写项失败: %w", err)
			}
			if len(newFields) > 0 {
				if err := tx.Create(&newFields).Error; err != nil {
					return fmt.Errorf("保存下单填写项失败: %w", err)
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err := s.reloadProduct(ctx, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// reloadProduct 重新读取商品（含两个关联），保证返回最新数据。
func (s *ProductService) reloadProduct(ctx context.Context, p *model.Product) error {
	if err := s.db.WithContext(ctx).
		Preload("Denominations", orderDenominations).
		Preload("Fields", orderFields).
		First(p, p.ID).Error; err != nil {
		return fmt.Errorf("刷新商品数据失败: %w", err)
	}
	return nil
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
