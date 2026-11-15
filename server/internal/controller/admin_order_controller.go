package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// AdminOrderController 运营后台订单管理接口（/api/v1/admin/orders/*）。
type AdminOrderController struct {
	svc *service.AdminOrderService
}

func NewAdminOrderController(svc *service.AdminOrderService) *AdminOrderController {
	return &AdminOrderController{svc: svc}
}

// List 订单分页列表 GET /api/v1/admin/orders
func (ctl *AdminOrderController) List(c *gin.Context) {
	var req service.ListAdminOrderRequest
	if !bindQuery(c, &req) {
		return
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	req.Page, req.PageSize = page, pageSize
	list, total, err := ctl.svc.List(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Page(c, list, total)
}

// Get 订单详情 GET /api/v1/admin/orders/:id
func (ctl *AdminOrderController) Get(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	v, err := ctl.svc.Get(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, v)
}

// Update 编辑订单 PUT /api/v1/admin/orders/:id
func (ctl *AdminOrderController) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req service.UpdateAdminOrderRequest
	if !bindJSON(c, &req) {
		return
	}
	v, err := ctl.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, v)
}
