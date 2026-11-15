package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// OrderController C 端下单接口（需用户登录）。
type OrderController struct {
	svc *service.OrderService
}

func NewOrderController(svc *service.OrderService) *OrderController {
	return &OrderController{svc: svc}
}

// Create 创建订单 POST /api/v1/orders
func (ctl *OrderController) Create(c *gin.Context) {
	var req service.CreateOrderRequest
	if !bindJSON(c, &req) {
		return
	}
	v, err := ctl.svc.Create(c.Request.Context(), c.GetUint64("user_id"), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, v)
}
