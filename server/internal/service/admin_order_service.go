package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"niulai/internal/model"
)

// AdminOrderService 运营后台订单管理（只读：列表与详情）。
// 与 C 端公开视图不同，后台视图包含渠道成本、利润等内部经营数据。
type AdminOrderService struct {
	db *gorm.DB
}

func NewAdminOrderService(db *gorm.DB) *AdminOrderService {
	return &AdminOrderService{db: db}
}

// ListAdminOrderRequest 后台订单分页查询。keyword 匹配订单号/充值账号/用户账号/商品名。
type ListAdminOrderRequest struct {
	Status   *int8  `form:"status"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// AdminOrderView 后台订单视图（含内部经营数据，仅 admin 接口使用）。
type AdminOrderView struct {
	ID             uint64               `gorm:"column:id"              json:"id"`
	OrderNo        string               `gorm:"column:order_no"        json:"order_no"`
	UserID         uint64               `gorm:"column:user_id"         json:"user_id"`
	UserPhone      string               `gorm:"column:user_phone"      json:"user_phone"`
	UserNickname   string               `gorm:"column:user_nickname"   json:"user_nickname"`
	ProductID      uint64               `gorm:"column:product_id"      json:"product_id"`
	ProductName    string               `gorm:"column:product_name"    json:"product_name"`
	ChannelID      uint64               `gorm:"column:channel_id"      json:"channel_id"`
	ChannelName    string               `gorm:"column:channel_name"    json:"channel_name"`
	Amount         model.Money          `gorm:"column:amount"          json:"amount"`
	ReceivedAmount model.Money          `gorm:"column:received_amount" json:"received_amount"`
	CostAmount     model.Money          `gorm:"column:cost_amount"     json:"cost_amount"`
	Profit         model.Money          `gorm:"column:profit"          json:"profit"`
	Status         int8                 `gorm:"column:status"          json:"status"`
	Account        string               `gorm:"column:account"         json:"account"`
	FormData       string               `gorm:"column:form_data"       json:"-"`
	Fields         []OrderFormFieldView `gorm:"-"                      json:"fields,omitempty"`
	CallbackAt     *time.Time           `gorm:"column:callback_at"     json:"callback_at"`
	CreatedAt      time.Time            `gorm:"column:created_at"      json:"created_at"`
}

// adminOrderSelect 列表与详情共用的投影列，关联名称统一用 COALESCE 兜底。
const adminOrderSelect = `o.id, o.order_no, o.user_id,
	COALESCE(u.phone, '') AS user_phone, COALESCE(u.nickname, '') AS user_nickname,
	o.product_id, COALESCE(p.name, '') AS product_name,
	o.channel_id, COALESCE(ch.name, '') AS channel_name,
	o.amount, o.received_amount, o.cost_amount, o.profit,
	o.status, o.account, o.form_data, o.callback_at, o.created_at`

func (s *AdminOrderService) base(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).Table("orders o").
		Joins("LEFT JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN channels ch ON ch.id = o.channel_id AND ch.deleted_at IS NULL").
		Where("o.deleted_at IS NULL")
}

func (s *AdminOrderService) filtered(ctx context.Context, req *ListAdminOrderRequest) *gorm.DB {
	q := s.base(ctx)
	if req.Status != nil {
		q = q.Where("o.status = ?", *req.Status)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("o.order_no LIKE ? OR o.account LIKE ? OR u.phone LIKE ? OR p.name LIKE ?",
			like, like, like, like)
	}
	return q
}

// List 分页查询订单，按 ID 倒序。
func (s *AdminOrderService) List(ctx context.Context, req *ListAdminOrderRequest) ([]AdminOrderView, int64, error) {
	var total int64
	if err := s.filtered(ctx, req).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计订单数量失败: %w", err)
	}
	list := make([]AdminOrderView, 0)
	if err := s.filtered(ctx, req).
		Select(adminOrderSelect).
		Order("o.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询订单列表失败: %w", err)
	}
	return list, total, nil
}

// Get 订单详情，附带下单填写项快照；查不到返回 ErrOrderNotFound。
func (s *AdminOrderService) Get(ctx context.Context, id uint64) (*AdminOrderView, error) {
	var v AdminOrderView
	if err := s.base(ctx).
		Select(adminOrderSelect).
		Where("o.id = ?", id).
		Scan(&v).Error; err != nil {
		return nil, fmt.Errorf("查询订单详情失败: %w", err)
	}
	// Scan 不返回 ErrRecordNotFound，查不到时是零值，需显式判断
	if v.ID == 0 {
		return nil, ErrOrderNotFound
	}
	v.Fields = parseFormData(v.FormData)
	return &v, nil
}
