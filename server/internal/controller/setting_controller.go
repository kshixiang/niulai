package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// SettingController 系统设置接口。
type SettingController struct {
	svc *service.SettingService
}

func NewSettingController(svc *service.SettingService) *SettingController {
	return &SettingController{svc: svc}
}

// Get 后台读取全部设置 GET /api/v1/admin/settings
func (ctl *SettingController) Get(c *gin.Context) {
	ctl.respondSettings(c)
}

// RegisterConfig 公开注册配置 GET /api/v1/register-config（注册页据此决定是否必填）
func (ctl *SettingController) RegisterConfig(c *gin.Context) {
	required, err := ctl.svc.IsRegisterCodeRequired(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"register_code_required": required})
}

// ExchangeRate 公开汇率 GET /api/v1/exchange-rate（H5 下单页据此把 CNY 应付换算为 USDT）
func (ctl *SettingController) ExchangeRate(c *gin.Context) {
	rate, err := ctl.svc.GetUSDTCNYRate(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"usdt_cny_rate": rate.StringFixed(2)})
}

// respondSettings 后台设置全量视图：注册开关、汇率与充值地址。
func (ctl *SettingController) respondSettings(c *gin.Context) {
	ctx := c.Request.Context()
	required, err := ctl.svc.IsRegisterCodeRequired(ctx)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	rate, err := ctl.svc.GetUSDTCNYRate(ctx)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	address, err := ctl.svc.GetRechargeAddress(ctx)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{
		"register_code_required": required,
		"usdt_cny_rate":          rate.StringFixed(2),
		"recharge_trc20_address": address,
	})
}

// UpdateSettingRequest 后台更新设置请求。各字段均可选，至少提供一项；
// 未提供的字段保持不变，兼容旧版只更新注册开关的调用。
type UpdateSettingRequest struct {
	RegisterCodeRequired *bool   `json:"register_code_required"`
	USDTCNYRate          *string `json:"usdt_cny_rate"`
	RechargeAddress      *string `json:"recharge_trc20_address"`
}

// Update 后台更新设置 PUT /api/v1/admin/settings
func (ctl *SettingController) Update(c *gin.Context) {
	var req UpdateSettingRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.RegisterCodeRequired == nil && req.USDTCNYRate == nil && req.RechargeAddress == nil {
		response.Fail(c, response.CodeInvalidParam, "未提供要更新的设置项")
		return
	}
	ctx := c.Request.Context()
	if req.RechargeAddress != nil {
		if err := ctl.svc.SetRechargeAddress(ctx, *req.RechargeAddress); err != nil {
			handleServiceError(c, err)
			return
		}
	}
	if req.RegisterCodeRequired != nil {
		if err := ctl.svc.SetRegisterCodeRequired(ctx, *req.RegisterCodeRequired); err != nil {
			handleServiceError(c, err)
			return
		}
	}
	if req.USDTCNYRate != nil {
		if err := ctl.svc.SetUSDTCNYRate(ctx, *req.USDTCNYRate); err != nil {
			handleServiceError(c, err)
			return
		}
	}
	ctl.respondSettings(c)
}
