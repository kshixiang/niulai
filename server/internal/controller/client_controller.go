package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// ClientController C 端公开查询接口（/api/v1/*，无需鉴权）。
type ClientController struct {
	svc *service.ClientService
}

func NewClientController(svc *service.ClientService) *ClientController {
	return &ClientController{svc: svc}
}

// ListProducts 在售商品列表 GET /api/v1/products
func (ctl *ClientController) ListProducts(c *gin.Context) {
	var req service.ListClientProductRequest
	if !bindQuery(c, &req) {
		return
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	req.Page, req.PageSize = page, pageSize

	list, total, err := ctl.svc.ListProducts(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Page(c, list, total)
}

// GetProduct 商品详情 GET /api/v1/products/:id（供 H5 下单页）
func (ctl *ClientController) GetProduct(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	v, err := ctl.svc.GetProduct(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, v)
}

// ListOrders 订单列表 GET /api/v1/orders
func (ctl *ClientController) ListOrders(c *gin.Context) {
	var req service.ListClientOrderRequest
	if !bindQuery(c, &req) {
		return
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	req.Page, req.PageSize = page, pageSize

	list, total, err := ctl.svc.ListOrders(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Page(c, list, total)
}

// OrderCounts 订单各状态计数 GET /api/v1/orders/counts
func (ctl *ClientController) OrderCounts(c *gin.Context) {
	counts, total, err := ctl.svc.OrderCounts(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"counts": counts, "total": total})
}

// GetOrder 订单详情 GET /api/v1/orders/:id
func (ctl *ClientController) GetOrder(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	v, err := ctl.svc.GetOrder(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, v)
}

// ListAgents 代理列表 GET /api/v1/agents
func (ctl *ClientController) ListAgents(c *gin.Context) {
	var req service.ListClientAgentRequest
	if !bindQuery(c, &req) {
		return
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	req.Page, req.PageSize = page, pageSize

	list, total, err := ctl.svc.ListAgents(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Page(c, list, total)
}
