package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// ChannelController 渠道管理接口
type ChannelController struct {
	svc *service.ChannelService
}

func NewChannelController(svc *service.ChannelService) *ChannelController {
	return &ChannelController{svc: svc}
}

// Create 新增渠道 POST /api/v1/admin/channels
func (ctl *ChannelController) Create(c *gin.Context) {
	var req service.CreateChannelRequest
	if !bindJSON(c, &req) {
		return
	}
	ch, err := ctl.svc.Create(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, ch)
}

// List 分页查询渠道 GET /api/v1/admin/channels
func (ctl *ChannelController) List(c *gin.Context) {
	var req service.ListChannelRequest
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

// Update 更新渠道 PUT /api/v1/admin/channels/:id
func (ctl *ChannelController) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req service.UpdateChannelRequest
	if !bindJSON(c, &req) {
		return
	}
	ch, err := ctl.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, ch)
}
