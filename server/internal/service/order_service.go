package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"niulai/internal/model"
)

// 下单相关业务错误
var (
	// ErrProductOffline 商品未上架，不允许下单
	ErrProductOffline = errors.New("商品已下架")
	// ErrInvalidOrderParam 下单参数不合法（金额、填写项等）
	ErrInvalidOrderParam = errors.New("下单参数不合法")
)

// CreateOrderRequest 创建订单请求。
// 实付金额一律由服务端按商品折扣与面值重算，前端展示的售价不参与计算。
type CreateOrderRequest struct {
	ProductID      uint64            `json:"product_id" binding:"required,min=1"`
	DenominationID uint64            `json:"denomination_id"` // 选预设面值时传，与 FaceValue 二选一
	FaceValue      string            `json:"face_value"`      // 自定义面值时传
	Fields         map[string]string `json:"fields"`          // 动态填写项 field_key -> value
	Account        string            `json:"account"`         // 主充值账号
}

// CreateOrderView 下单结果
type CreateOrderView struct {
	ID      uint64      `json:"id"`
	OrderNo string      `json:"order_no"`
	Amount  model.Money `json:"amount"`
	Status  int8        `json:"status"`
}

// orderFormItem 下单填写项快照：随订单落库，避免后续改配置影响历史订单。
type orderFormItem struct {
	FieldKey string `json:"field_key"`
	Label    string `json:"label"`
	Value    string `json:"value"`
}

// OrderService C 端下单服务。
type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

// Create 创建订单：校验商品与面值、按后台折扣重算金额、校验填写项后落库。
func (s *OrderService) Create(ctx context.Context, userID uint64, req *CreateOrderRequest) (*CreateOrderView, error) {
	var p model.Product
	if err := s.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", req.ProductID).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("查询商品失败: %w", err)
	}
	if p.Status != model.ProductStatusOnline {
		return nil, ErrProductOffline
	}

	face, err := s.resolveFaceValue(ctx, req)
	if err != nil {
		return nil, err
	}
	// 金额全部由服务端计算：售价=面值×出售折扣÷10，成本=面值×渠道折扣÷10
	amount := faceToPrice(face, p.Discount)
	cost := faceToPrice(face, p.ChannelDiscount)

	formData, account, err := s.buildFormData(ctx, req.ProductID, req.Fields, req.Account)
	if err != nil {
		return nil, err
	}

	order := &model.Order{
		OrderNo:        genOrderNo(),
		UserID:         userID,
		ProductID:      p.ID,
		ChannelID:      p.ChannelID,
		Amount:         amount,
		ReceivedAmount: decimal.Zero,
		CostAmount:     cost,
		Profit:         amount.Sub(cost),
		Status:         model.OrderStatusPending,
		Account:        account,
		FormData:       formData,
	}
	if err := s.db.WithContext(ctx).Create(order).Error; err != nil {
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}
	return &CreateOrderView{
		ID:      order.ID,
		OrderNo: order.OrderNo,
		Amount:  order.Amount,
		Status:  order.Status,
	}, nil
}

// resolveFaceValue 取本次下单面值：预设档从库里读（不信任前端金额），自定义档校验输入。
func (s *OrderService) resolveFaceValue(ctx context.Context, req *CreateOrderRequest) (model.Money, error) {
	if req.DenominationID > 0 {
		var d model.ProductDenomination
		if err := s.db.WithContext(ctx).
			Where("id = ? AND product_id = ? AND deleted_at IS NULL", req.DenominationID, req.ProductID).
			First(&d).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return decimal.Decimal{}, fmt.Errorf("%w: 价格类别不存在", ErrInvalidOrderParam)
			}
			return decimal.Decimal{}, fmt.Errorf("查询价格类别失败: %w", err)
		}
		if !d.FaceValue.IsPositive() {
			return decimal.Decimal{}, fmt.Errorf("%w: 价格类别面值异常", ErrInvalidOrderParam)
		}
		return d.FaceValue, nil
	}

	raw := strings.TrimSpace(req.FaceValue)
	if raw == "" {
		return decimal.Decimal{}, fmt.Errorf("%w: 请选择或输入订单金额", ErrInvalidOrderParam)
	}
	face, err := decimal.NewFromString(raw)
	if err != nil || !face.IsPositive() || !face.Equal(face.Round(2)) ||
		face.GreaterThan(decimal.NewFromInt(1000000)) {
		return decimal.Decimal{}, fmt.Errorf("%w: 订单金额须为大于0且不超过2位小数的数值", ErrInvalidOrderParam)
	}
	return face, nil
}

// buildFormData 按商品配置校验并归一化填写项，返回 JSON 快照与主充值账号。
// 只落库后台配置的字段，前端多传的键一律忽略。
func (s *OrderService) buildFormData(ctx context.Context, productID uint64, input map[string]string, account string) (string, string, error) {
	var fields []model.ProductOrderField
	if err := s.db.WithContext(ctx).
		Where("product_id = ? AND deleted_at IS NULL", productID).
		Order("sort ASC, id ASC").
		Find(&fields).Error; err != nil {
		return "", "", fmt.Errorf("查询下单填写项失败: %w", err)
	}

	items := make([]orderFormItem, 0, len(fields))
	for _, f := range fields {
		v := strings.TrimSpace(input[f.FieldKey])
		if f.Required == 1 && v == "" {
			return "", "", fmt.Errorf("%w: 请填写%s", ErrInvalidOrderParam, f.Label)
		}
		if v == "" {
			continue
		}
		if len([]rune(v)) > 128 {
			return "", "", fmt.Errorf("%w: %s不能超过128字", ErrInvalidOrderParam, f.Label)
		}
		switch f.Type {
		case OrderFieldTypeNumber:
			if _, err := decimal.NewFromString(v); err != nil {
				return "", "", fmt.Errorf("%w: %s须为数字", ErrInvalidOrderParam, f.Label)
			}
		case OrderFieldTypeSelect:
			if !containsString(parseOptions(f.Options), v) {
				return "", "", fmt.Errorf("%w: %s取值不合法", ErrInvalidOrderParam, f.Label)
			}
		}
		items = append(items, orderFormItem{FieldKey: f.FieldKey, Label: f.Label, Value: v})
		if account == "" {
			account = v // 未显式指定账号时，按配置顺序取第一个填写项
		}
	}

	data, err := json.Marshal(items)
	if err != nil {
		return "", "", fmt.Errorf("序列化下单信息失败: %w", err)
	}
	account = strings.TrimSpace(account)
	if len([]rune(account)) > 128 {
		return "", "", fmt.Errorf("%w: 充值账号过长", ErrInvalidOrderParam)
	}
	return string(data), account, nil
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// genOrderNo 生成订单号：NL + 秒级时间 + 6 位随机数，长度小于 varchar(64)。
func genOrderNo() string {
	return fmt.Sprintf("NL%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}
