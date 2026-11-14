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

// Get 后台读取设置 GET /api/v1/admin/settings
func (ctl *SettingController) Get(c *gin.Context) {
	ctl.respondRegisterConfig(c)
}

// RegisterConfig 公开注册配置 GET /api/v1/register-config（注册页据此决定是否必填）
func (ctl *SettingController) RegisterConfig(c *gin.Context) {
	ctl.respondRegisterConfig(c)
}

func (ctl *SettingController) respondRegisterConfig(c *gin.Context) {
	required, err := ctl.svc.IsRegisterCodeRequired(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"register_code_required": required})
}

// UpdateSettingRequest 后台更新设置请求
type UpdateSettingRequest struct {
	RegisterCodeRequired *bool `json:"register_code_required" binding:"required"`
}

// Update 后台更新设置 PUT /api/v1/admin/settings
func (ctl *SettingController) Update(c *gin.Context) {
	var req UpdateSettingRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := ctl.svc.SetRegisterCodeRequired(c.Request.Context(), *req.RegisterCodeRequired); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"register_code_required": *req.RegisterCodeRequired})
}
