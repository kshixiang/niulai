package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// AdminUserController 运营后台用户管理接口（/api/v1/admin/users/*）。
type AdminUserController struct {
	svc *service.AdminUserService
}

func NewAdminUserController(svc *service.AdminUserService) *AdminUserController {
	return &AdminUserController{svc: svc}
}

// List 分页查询已注册用户 GET /api/v1/admin/users
func (ctl *AdminUserController) List(c *gin.Context) {
	var req service.ListUserRequest
	if !bindQuery(c, &req) {
		return
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	req.Page = page
	req.PageSize = pageSize
	list, total, err := ctl.svc.List(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Page(c, list, total)
}

// UpdateStatus 启用/封禁切换 PUT /api/v1/admin/users/:id/status
func (ctl *AdminUserController) UpdateStatus(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req service.UpdateUserStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	view, err := ctl.svc.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, view)
}
