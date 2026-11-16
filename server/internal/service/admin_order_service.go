package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"niulai/internal/model"
)

// AdminOrderService 运营后台订单管理（列表、详情与编辑）。
// 与 C 端公开视图不同，后台视图包含渠道成本、利润等内部经营数据。
type AdminOrderService struct {
	db *gorm.DB
}

func NewAdminOrderService(db *gorm.DB) *AdminOrderService {
	return &AdminOrderService{db: db}
}

// ErrOrderFundConflict 订单资金状态已变更（重复结算或非法跃迁），需刷新重试。
var ErrOrderFundConflict = errors.New("订单资金状态已变更，请刷新重试")

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
	FrozenUSDT     model.Money          `gorm:"column:frozen_usdt"     json:"frozen_usdt"`
	UsdtCnyRate    model.Money          `gorm:"column:usdt_cny_rate"   json:"usdt_cny_rate"`
	FundStatus     int8                 `gorm:"column:fund_status"     json:"fund_status"`
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
	o.frozen_usdt, o.usdt_cny_rate, o.fund_status,
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

// UpdateAdminOrderRequest 后台编辑订单请求。
// 订单号、商品、用户、渠道为订单身份，不允许修改；
// 实付/成本/已到账可修正，利润由服务端按 实付-成本 自动重算，不接受前端传入。
type UpdateAdminOrderRequest struct {
	Account        *string `json:"account"`
	Status         *int8   `json:"status"`
	Amount         *string `json:"amount"`
	ReceivedAmount *string `json:"received_amount"`
	CostAmount     *string `json:"cost_amount"`
}

// maxOrderAmount 编辑时金额上限，避免超出 decimal(10,2) 列范围。
var maxOrderAmount = decimal.NewFromInt(10000000)

func parseOrderAmount(s string) (model.Money, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return decimal.Decimal{}, fmt.Errorf("%w: 金额不能为空", ErrInvalidOrderParam)
	}
	d, err := decimal.NewFromString(raw)
	if err != nil || d.IsNegative() || !d.Equal(d.Round(2)) || d.GreaterThan(maxOrderAmount) {
		return decimal.Decimal{}, fmt.Errorf("%w: 金额须为不小于0且不超过2位小数的数值", ErrInvalidOrderParam)
	}
	return d, nil
}

// Update 编辑订单：仅更新显式传入的字段，返回更新后的详情视图。
func (s *AdminOrderService) Update(ctx context.Context, id uint64, req *UpdateAdminOrderRequest) (*AdminOrderView, error) {
	var o model.Order
	if err := s.db.WithContext(ctx).First(&o, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	updates := map[string]interface{}{}
	if req.Account != nil {
		account := strings.TrimSpace(*req.Account)
		if len([]rune(account)) > 128 {
			return nil, fmt.Errorf("%w: 充值账号不能超过128字", ErrInvalidOrderParam)
		}
		updates["account"] = account
	}
	if req.Status != nil {
		if *req.Status < model.OrderStatusPending || *req.Status > model.OrderStatusRefunded {
			return nil, fmt.Errorf("%w: 订单状态不合法", ErrInvalidOrderParam)
		}
		updates["status"] = *req.Status
	}
	if req.Amount != nil {
		v, err := parseOrderAmount(*req.Amount)
		if err != nil {
			return nil, err
		}
		updates["amount"] = v
	}
	if req.ReceivedAmount != nil {
		v, err := parseOrderAmount(*req.ReceivedAmount)
		if err != nil {
			return nil, err
		}
		updates["received_amount"] = v
	}
	if req.CostAmount != nil {
		v, err := parseOrderAmount(*req.CostAmount)
		if err != nil {
			return nil, err
		}
		updates["cost_amount"] = v
	}

	// 利润始终按最终 实付-渠道成本 重算，避免手改金额后账目对不上
	finalAmount, finalCost := o.Amount, o.CostAmount
	if v, ok := updates["amount"]; ok {
		finalAmount = v.(model.Money)
	}
	if v, ok := updates["cost_amount"]; ok {
		finalCost = v.(model.Money)
	}
	updates["profit"] = finalAmount.Sub(finalCost)

	// 状态变更为结算态（成功/失败/退款）时触发资金动作，需在事务内与订单更新原子提交。
	var plan fundPlan
	var needFund bool
	if req.Status != nil && *req.Status != o.Status {
		var conflict bool
		plan, needFund, conflict = planFundTransition(*req.Status, o.FundStatus, o.FrozenUSDT)
		if conflict {
			return nil, ErrOrderFundConflict
		}
	}

	if needFund {
		updates["fund_status"] = plan.newFund
		if err := s.settleOrder(ctx, &o, updates, plan); err != nil {
			return nil, err
		}
	} else if err := s.db.WithContext(ctx).Model(&o).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新订单失败: %w", err)
	}
	return s.Get(ctx, id)
}

// fundPlan 描述一次订单结算所需的资金动作。
type fundPlan struct {
	newFund     int8                   // 目标资金状态
	walletUpd   map[string]interface{} // 钱包条件更新表达式（不含 version，由 settleOrder 追加）
	ledgerType  int8                   // 资金流水类型
	remark      string                 // 流水备注
	frozenGuard bool                   // 是否需要 frozen_balance >= 冻结额 守卫
}

// planFundTransition 根据目标订单状态与当前资金状态计算资金动作。
// ok=false 表示无需资金动作（幂等/非结算态/历史订单）；conflict=true 表示非法跃迁。
// X 取订单下单快照 FrozenUSDT，与事后手改的 amount 无关。
func planFundTransition(target, current int8, x model.Money) (p fundPlan, ok bool, conflict bool) {
	if !x.IsPositive() {
		// 历史订单未冻结金额，不触发资金动作
		return fundPlan{}, false, false
	}
	switch target {
	case model.OrderStatusSuccess: // 成功：冻结转实际扣款
		switch current {
		case model.OrderFundFrozen:
			return fundPlan{
				newFund:     model.OrderFundCharged,
				walletUpd:   map[string]interface{}{"frozen_balance": gorm.Expr("frozen_balance - ?", x)},
				ledgerType:  model.WalletTxDeduct,
				remark:      "订单成功扣款",
				frozenGuard: true,
			}, true, false
		case model.OrderFundCharged: // 已扣款，幂等
			return fundPlan{}, false, false
		default: // 已退还不能再扣款
			return fundPlan{}, false, true
		}
	case model.OrderStatusFailed, model.OrderStatusRefunded: // 失败/退款：归还
		switch current {
		case model.OrderFundFrozen: // 冻结直接解冻归还
			ledgerType, remark := int8(model.WalletTxRefund), "订单退款归还"
			if target == model.OrderStatusFailed {
				ledgerType, remark = model.WalletTxUnfreezeFailed, "订单失败解冻归还"
			}
			return fundPlan{
				newFund: model.OrderFundReturned,
				walletUpd: map[string]interface{}{
					"frozen_balance": gorm.Expr("frozen_balance - ?", x),
					"balance":        gorm.Expr("balance + ?", x),
				},
				ledgerType:  ledgerType,
				remark:      remark,
				frozenGuard: true,
			}, true, false
		case model.OrderFundCharged: // 已扣款后退款：冻结已归零，只加可用余额
			return fundPlan{
				newFund:    model.OrderFundReturned,
				walletUpd:  map[string]interface{}{"balance": gorm.Expr("balance + ?", x)},
				ledgerType: model.WalletTxRefund,
				remark:     "已扣款订单退款归还",
			}, true, false
		default: // 已退还，幂等
			return fundPlan{}, false, false
		}
	}
	// 非结算态（待支付/已支付/充值中）不触发资金动作
	return fundPlan{}, false, false
}

// settleOrder 在单事务内完成：订单资金状态抢占（幂等闸）+ 钱包条件更新 + 资金流水。
// 订单以 fund_status 旧值作条件更新，并发/重复结算只有一方 RowsAffected>0，避免重复出金。
func (s *AdminOrderService) settleOrder(ctx context.Context, o *model.Order, updates map[string]interface{}, plan fundPlan) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 抢占订单资金状态：仅当 fund_status 仍为读到的旧值时才更新
		res := tx.Model(&model.Order{}).
			Where("id = ? AND fund_status = ? AND deleted_at IS NULL", o.ID, o.FundStatus).
			Updates(updates)
		if res.Error != nil {
			return fmt.Errorf("更新订单失败: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return ErrOrderFundConflict
		}

		// 2) 钱包条件更新：版本守卫（必要时加冻结余额守卫）
		var w model.Wallet
		if err := tx.Where("user_id = ?", o.UserID).First(&w).Error; err != nil {
			return fmt.Errorf("查询资金账户失败: %w", err)
		}
		where := "user_id = ? AND version = ?"
		args := []interface{}{o.UserID, w.Version}
		if plan.frozenGuard {
			where += " AND frozen_balance >= ?"
			args = append(args, o.FrozenUSDT)
		}
		walletUpd := make(map[string]interface{}, len(plan.walletUpd)+1)
		for k, v := range plan.walletUpd {
			walletUpd[k] = v
		}
		walletUpd["version"] = gorm.Expr("version + 1")
		wr := tx.Model(&model.Wallet{}).Where(where, args...).Updates(walletUpd)
		if wr.Error != nil {
			return fmt.Errorf("更新资金账户失败: %w", wr.Error)
		}
		if wr.RowsAffected == 0 {
			return ErrOrderFundConflict
		}

		// 3) 读变动后余额并写资金流水
		var after model.Wallet
		if err := tx.Where("user_id = ?", o.UserID).First(&after).Error; err != nil {
			return fmt.Errorf("查询变动后资金账户失败: %w", err)
		}
		txn := model.WalletTransaction{
			UserID:        o.UserID,
			OrderID:       o.ID,
			Type:          plan.ledgerType,
			Amount:        o.FrozenUSDT,
			BalanceBefore: w.Balance,
			BalanceAfter:  after.Balance,
			FrozenBefore:  w.FrozenBalance,
			FrozenAfter:   after.FrozenBalance,
			Ref:           o.OrderNo,
			Remark:        plan.remark,
		}
		if err := tx.Create(&txn).Error; err != nil {
			return fmt.Errorf("记录资金流水失败: %w", err)
		}
		return nil
	})
}
