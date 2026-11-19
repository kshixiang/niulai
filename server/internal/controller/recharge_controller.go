package controller

import (
	"github.com/gin-gonic/gin"
	"niulai/internal/service"
	"niulai/pkg/response"
)

type RechargeController struct{ svc *service.RechargeService }

func NewRechargeController(svc *service.RechargeService) *RechargeController {
	return &RechargeController{svc: svc}
}

func (ctl *RechargeController) Config(c *gin.Context) {
	data, err := ctl.svc.Config(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, data)
}

func (ctl *RechargeController) Create(c *gin.Context) {
	var req service.CreateRechargeRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := ctl.svc.Create(c.Request.Context(), c.GetUint64("user_id"), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, data)
}

func (ctl *RechargeController) List(c *gin.Context)      { ctl.list(c, c.GetUint64("user_id")) }
func (ctl *RechargeController) AdminList(c *gin.Context) { ctl.list(c, 0) }
func (ctl *RechargeController) list(c *gin.Context, userID uint64) {
	var req service.ListRechargeRequest
	if !bindQuery(c, &req) {
		return
	}
	req.Page, req.PageSize = normalizePage(req.Page, req.PageSize)
	data, total, err := ctl.svc.List(c.Request.Context(), userID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Page(c, data, total)
}

func (ctl *RechargeController) Review(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req service.ReviewRechargeRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := ctl.svc.Review(c.Request.Context(), id, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, data)
}
