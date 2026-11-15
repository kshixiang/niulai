package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// WalletController C 端资金账户接口（/api/v1/wallet）。
type WalletController struct {
	svc *service.WalletService
}

func NewWalletController(svc *service.WalletService) *WalletController {
	return &WalletController{svc: svc}
}

// Get 当前用户资金账户 GET /api/v1/wallet
func (ctl *WalletController) Get(c *gin.Context) {
	view, err := ctl.svc.GetByUser(c.Request.Context(), c.GetUint64("user_id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, view)
}
