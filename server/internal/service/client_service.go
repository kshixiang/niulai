package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"niulai/internal/model"
)

// ClientService C 端公开查询服务（/api/v1/*，无需鉴权）。
//
// 只投影可对外展示的字段：成本价、利润、上游渠道、用户密码等内部数据
// 一律不出现在公开视图中，避免通过接口泄露经营数据。
type ClientService struct {
	db *gorm.DB
}

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
	Status   *int8 `form:"status"`
	Page     int   `form:"page"`
	PageSize int   `form:"page_size"`
}

// ListClientAgentRequest 代理查询
type ListClientAgentRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ---------- 公开视图 ----------

// ClientProductView 商品对外视图
type ClientProductView struct {
	ID            uint64      `gorm:"column:id"             json:"id"`
	Name          string      `gorm:"column:name"           json:"name"`
	Type          string      `gorm:"column:type"           json:"type"`
	CoverImage    string      `gorm:"column:cover_image"    json:"cover_image"`
	SalePrice     model.Money `gorm:"column:sale_price"     json:"sale_price"`
	DiscountLabel string      `gorm:"column:discount_label" json:"discount_label"`
	Description   string      `gorm:"column:description"    json:"description"`
}

// ClientOrderView 订单对外视图
type ClientOrderView struct {
	ID          uint64      `gorm:"column:id"           json:"id"`
	OrderNo     string      `gorm:"column:order_no"     json:"order_no"`
	ProductName string      `gorm:"column:product_name" json:"product_name"`
	Amount      model.Money `gorm:"column:amount"       json:"amount"`
	Status      int8        `gorm:"column:status"       json:"status"`
	Account     string      `gorm:"column:account"      json:"account"`
	CreatedAt   time.Time   `gorm:"column:created_at"   json:"created_at"`
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

// ListProducts 分页查询上架商品，支持按类型过滤。
// Page / PageSize 由 controller 标准化后传入。
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
		Select("id, name, type, cover_image, sale_price, discount_label, description").
		Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询在售商品失败: %w", err)
	}
	return list, total, nil
}

// ListOrders 分页查询订单，关联商品名便于展示；支持按状态过滤。
func (s *ClientService) ListOrders(ctx context.Context, req *ListClientOrderRequest) ([]ClientOrderView, int64, error) {
	base := func() *gorm.DB {
		q := s.db.WithContext(ctx).Table("orders o").
			Joins("LEFT JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL").
			Where("o.deleted_at IS NULL")
		if req.Status != nil {
			q = q.Where("o.status = ?", *req.Status)
		}
		return q
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计订单失败: %w", err)
	}

	list := make([]ClientOrderView, 0)
	if err := base().
		Select("o.id, o.order_no, COALESCE(p.name, '') AS product_name, o.amount, o.status, o.account, o.created_at").
		Order("o.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询订单失败: %w", err)
	}
	return list, total, nil
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
