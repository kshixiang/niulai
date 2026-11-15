package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"niulai/internal/model"
)

// ClientService C 端公开查询服务（/api/v1/*，无需鉴权）。
//
// 只投影可对外展示的字段：渠道折扣、渠道价、盈利、成本价、上游渠道、用户密码等内部数据
// 一律不出现在公开视图中，避免通过接口泄露经营数据。
type ClientService struct {
	db *gorm.DB
}

// 业务错误
var (
	ErrOrderNotFound = errors.New("订单不存在")
)

func NewClientService(db *gorm.DB) *ClientService {
	return &ClientService{db: db}
}

// ---------- 请求参数 ----------

// ListClientProductRequest 在售商品查询
type ListClientProductRequest struct {
	Type     string `form:"type"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// ListClientOrderRequest 订单查询
type ListClientOrderRequest struct {
	Status   *int8  `form:"status"`    // 状态过滤：0待接单 1已接单 2已完成 3已取消
	Keyword  string `form:"keyword"`   // 订单号 / 主充值号码 模糊搜索
	SearchBy string `form:"search_by"` // all(默认) | order_no | account
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// ListClientAgentRequest 代理查询
type ListClientAgentRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ---------- 公开视图 ----------

// ClientDenominationView 价格类别(面值)对外视图，售价(final_price)由服务端按出售折扣实时计算。
type ClientDenominationView struct {
	ID         uint64      `json:"id"`
	Name       string      `json:"name"`
	FaceValue  model.Money `json:"face_value"`
	FinalPrice model.Money `json:"final_price"`
	Sort       int         `json:"sort"`
}

// ClientOrderFieldView 下单填写项对外视图，不含 id/product_id 等内部字段。
type ClientOrderFieldView struct {
	FieldKey    string   `json:"field_key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    int8     `json:"required"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options"`
	Sort        int      `json:"sort"`
}

// ClientProductView 商品列表对外视图（首页只展示折扣角标，不返回价格条目）。
type ClientProductView struct {
	ID          uint64      `gorm:"column:id"           json:"id"`
	Name        string      `gorm:"column:name"         json:"name"`
	Type        string      `gorm:"column:type"         json:"type"`
	CoverImage  string      `gorm:"column:cover_image"  json:"cover_image"`
	Discount    model.Money `gorm:"column:discount"     json:"discount"`
	Description string      `gorm:"column:description"  json:"description"`
}

// ClientProductDetailView 商品详情对外视图（供 H5 下单页）。
type ClientProductDetailView struct {
	ID            uint64                   `gorm:"column:id"           json:"id"`
	Name          string                   `gorm:"column:name"         json:"name"`
	Type          string                   `gorm:"column:type"         json:"type"`
	CoverImage    string                   `gorm:"column:cover_image"  json:"cover_image"`
	Discount      model.Money              `gorm:"column:discount"     json:"discount"`
	Description   string                   `gorm:"column:description"  json:"description"`
	Denominations []ClientDenominationView `gorm:"-" json:"denominations"`
	Fields        []ClientOrderFieldView   `gorm:"-" json:"fields"`
}

// OrderFormFieldView 订单填写项对外视图，仅暴露中文名与填写值。
type OrderFormFieldView struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// parseFormData 解析订单填写项快照（下单时落库的配置快照），仅对外暴露 label 与值。
func parseFormData(s string) []OrderFormFieldView {
	if s == "" {
		return nil
	}
	var items []orderFormItem
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		return nil
	}
	out := make([]OrderFormFieldView, 0, len(items))
	for _, it := range items {
		out = append(out, OrderFormFieldView{Label: it.Label, Value: it.Value})
	}
	return out
}

// ClientOrderView 订单对外视图
type ClientOrderView struct {
	ID              uint64               `gorm:"column:id"             json:"id"`
	OrderNo         string               `gorm:"column:order_no"       json:"order_no"`
	ProductName     string               `gorm:"column:product_name"   json:"product_name"`
	Amount          model.Money          `gorm:"column:amount"         json:"amount"`
	ReceivedAmount  model.Money          `gorm:"column:received_amount" json:"received_amount"`
	RemainingAmount model.Money          `gorm:"-"                     json:"remaining_amount"` // 派生：订单金额 - 已到账金额
	Status          int8                 `gorm:"column:status"         json:"status"`
	Account         string               `gorm:"column:account"        json:"account"`
	FormData        string               `gorm:"column:form_data"      json:"-"`
	Fields          []OrderFormFieldView `gorm:"-"                     json:"fields,omitempty"`
	CreatedAt       time.Time            `gorm:"column:created_at"     json:"created_at"`
}

// ClientAgentView 代理对外视图
type ClientAgentView struct {
	ID         uint64    `gorm:"column:id"          json:"id"`
	UserID     uint64    `gorm:"column:user_id"     json:"user_id"`
	Level      int8      `gorm:"column:level"       json:"level"`
	InviteCode string    `gorm:"column:invite_code" json:"invite_code"`
	CreatedAt  time.Time `gorm:"column:created_at"  json:"created_at"`
}

// ---------- 查询方法 ----------

// ListProducts 分页查询上架商品，支持按类型过滤。首页只展示折扣角标，不返回面额与价格。
func (s *ClientService) ListProducts(ctx context.Context, req *ListClientProductRequest) ([]ClientProductView, int64, error) {
	base := func() *gorm.DB {
		q := s.db.WithContext(ctx).Table("products").
			Where("status = ?", model.ProductStatusOnline).
			Where("deleted_at IS NULL")
		if req.Type != "" {
			q = q.Where("type = ?", req.Type)
		}
		return q
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计在售商品失败: %w", err)
	}

	list := make([]ClientProductView, 0)
	if err := base().
		Select("id, name, type, cover_image, discount, description").
		Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询在售商品失败: %w", err)
	}
	return list, total, nil
}

// GetProduct 公开商品详情：仅上架、未软删商品，只返回可对外展示的字段。
// 供 H5 下单页使用；绝不暴露 channel_discount/成本/盈利/渠道等内部数据。
func (s *ClientService) GetProduct(ctx context.Context, id uint64) (*ClientProductDetailView, error) {
	var v ClientProductDetailView
	if err := s.db.WithContext(ctx).Table("products").
		Select("id, name, type, cover_image, discount, description").
		Where("id = ? AND status = ? AND deleted_at IS NULL", id, model.ProductStatusOnline).
		Scan(&v).Error; err != nil {
		return nil, fmt.Errorf("查询商品详情失败: %w", err)
	}
	// Scan 不返回 ErrRecordNotFound，查不到时是零值，需显式判断
	if v.ID == 0 {
		return nil, ErrProductNotFound
	}

	denoms, err := s.denominationViews(ctx, id, v.Discount)
	if err != nil {
		return nil, err
	}
	v.Denominations = denoms

	fields, err := s.fieldViews(ctx, id, v.Type)
	if err != nil {
		return nil, err
	}
	v.Fields = fields
	return &v, nil
}

// denominationViews 查询商品价格类别并计算售价(final_price)：面值 × 出售折扣 ÷ 10。
func (s *ClientService) denominationViews(ctx context.Context, productID uint64, discount model.Money) ([]ClientDenominationView, error) {
	var rows []model.ProductDenomination
	if err := s.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("sort ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询商品价格类别失败: %w", err)
	}
	out := make([]ClientDenominationView, 0, len(rows))
	for _, r := range rows {
		out = append(out, ClientDenominationView{
			ID:         r.ID,
			Name:       r.Name,
			FaceValue:  r.FaceValue,
			FinalPrice: faceToPrice(r.FaceValue, discount),
			Sort:       r.Sort,
		})
	}
	return out, nil
}

// fieldViews 查询商品下单填写项配置，options 解析为数组。
// 话费商品补齐后台未配置的手机号、用户名必填项，供 H5 下单页动态渲染。
func (s *ClientService) fieldViews(ctx context.Context, productID uint64, productType string) ([]ClientOrderFieldView, error) {
	var rows []model.ProductOrderField
	if err := s.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("sort ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询商品下单填写项失败: %w", err)
	}
	rows = appendPhoneFields(rows, productType)

	out := make([]ClientOrderFieldView, 0, len(rows))
	for _, r := range rows {
		out = append(out, ClientOrderFieldView{
			FieldKey:    r.FieldKey,
			Label:       r.Label,
			Type:        r.Type,
			Required:    r.Required,
			Placeholder: r.Placeholder,
			Options:     parseOptions(r.Options),
			Sort:        r.Sort,
		})
	}
	// 兜底项排在前面，按展示顺序重排 sort，避免与后台配置的旧序号冲突
	for i := range out {
		out[i].Sort = i
	}
	return out, nil
}

// ListOrders 分页查询订单，关联商品名便于展示；支持按状态、关键字（订单号/主充值号码）过滤。
func (s *ClientService) ListOrders(ctx context.Context, req *ListClientOrderRequest) ([]ClientOrderView, int64, error) {
	base := func() *gorm.DB {
		q := s.db.WithContext(ctx).Table("orders o").
			Joins("LEFT JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL").
			Where("o.deleted_at IS NULL")
		if req.Status != nil {
			q = q.Where("o.status = ?", *req.Status)
		}
		if kw := strings.TrimSpace(req.Keyword); kw != "" {
			like := "%" + kw + "%"
			switch req.SearchBy {
			case "order_no":
				q = q.Where("o.order_no LIKE ?", like)
			case "account":
				q = q.Where("o.account LIKE ?", like)
			default:
				q = q.Where("o.order_no LIKE ? OR o.account LIKE ?", like, like)
			}
		}
		return q
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计订单失败: %w", err)
	}

	list := make([]ClientOrderView, 0)
	if err := base().
		Select("o.id, o.order_no, COALESCE(p.name, '') AS product_name, o.amount, o.received_amount, o.status, o.account, o.created_at").
		Order("o.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询订单失败: %w", err)
	}
	for i := range list {
		list[i].RemainingAmount = list[i].Amount.Sub(list[i].ReceivedAmount)
	}
	return list, total, nil
}

// OrderCounts 统计各状态订单数量与总数，供订单中心状态 tab 角标展示。
func (s *ClientService) OrderCounts(ctx context.Context) (map[int8]int64, int64, error) {
	type row struct {
		Status int8  `gorm:"column:status"`
		Cnt    int64 `gorm:"column:cnt"`
	}
	var rows []row
	if err := s.db.WithContext(ctx).Table("orders o").
		Select("o.status AS status, COUNT(*) AS cnt").
		Where("o.deleted_at IS NULL").
		Group("o.status").
		Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("统计订单状态失败: %w", err)
	}
	counts := make(map[int8]int64)
	var total int64
	for _, r := range rows {
		counts[r.Status] = r.Cnt
		total += r.Cnt
	}
	return counts, total, nil
}

// GetOrder 订单详情（供订单详情页），关联商品名；查不到返回 ErrOrderNotFound。
func (s *ClientService) GetOrder(ctx context.Context, id uint64) (*ClientOrderView, error) {
	var v ClientOrderView
	if err := s.db.WithContext(ctx).Table("orders o").
		Joins("LEFT JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL").
		Select("o.id, o.order_no, COALESCE(p.name, '') AS product_name, o.amount, o.received_amount, o.status, o.account, o.form_data, o.created_at").
		Where("o.id = ? AND o.deleted_at IS NULL", id).
		Scan(&v).Error; err != nil {
		return nil, fmt.Errorf("查询订单详情失败: %w", err)
	}
	if v.ID == 0 {
		return nil, ErrOrderNotFound
	}
	v.RemainingAmount = v.Amount.Sub(v.ReceivedAmount)
	v.Fields = parseFormData(v.FormData)
	return &v, nil
}

// ListAgents 分页查询代理关系。
func (s *ClientService) ListAgents(ctx context.Context, req *ListClientAgentRequest) ([]ClientAgentView, int64, error) {
	base := func() *gorm.DB {
		return s.db.WithContext(ctx).Table("agents").Where("deleted_at IS NULL")
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计代理失败: %w", err)
	}

	list := make([]ClientAgentView, 0)
	if err := base().
		Select("id, user_id, level, invite_code, created_at").
		Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询代理失败: %w", err)
	}
	return list, total, nil
}
