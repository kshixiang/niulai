package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// RegisterCodeController 管理后台注册码接口。
type RegisterCodeController struct {
	svc *service.RegisterCodeService
}

func NewRegisterCodeController(svc *service.RegisterCodeService) *RegisterCodeController {
	return &RegisterCodeController{svc: svc}
}

// GenerateRegisterCodeRequest 批量生成请求
type GenerateRegisterCodeRequest struct {
	Count     int    `json:"count" binding:"required,min=1,max=100"`
	InviterID uint64 `json:"inviter_id"`
}

// Generate 批量生成 POST /api/v1/admin/register-codes
func (ctl *RegisterCodeController) Generate(c *gin.Context) {
	var req GenerateRegisterCodeRequest
	if !bindJSON(c, &req) {
		return
	}
	list, err := ctl.svc.Generate(c.Request.Context(), req.Count, req.InviterID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, list)
}

// List 分页列表 GET /api/v1/admin/register-codes
func (ctl *RegisterCodeController) List(c *gin.Context) {
	var req struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
	}
	if !bindQuery(c, &req) {
		return
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	list, total, err := ctl.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Page(c, list, total)
}

// UpdateRegisterCodeStatusRequest 启用/禁用请求
type UpdateRegisterCodeStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

// UpdateStatus 启用/禁用 PUT /api/v1/admin/register-codes/:id/status
func (ctl *RegisterCodeController) UpdateStatus(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req UpdateRegisterCodeStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := ctl.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "status": req.Status})
}
