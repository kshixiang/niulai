package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// ProductController 商品管理接口
type ProductController struct {
	svc *service.ProductService
}

func NewProductController(svc *service.ProductService) *ProductController {
	return &ProductController{svc: svc}
}

// Create 新增商品 POST /api/v1/admin/products
func (ctl *ProductController) Create(c *gin.Context) {
	var req service.CreateProductRequest
	if !bindJSON(c, &req) {
		return
	}
	p, err := ctl.svc.Create(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, p)
}

// List 分页查询商品 GET /api/v1/admin/products（支持 name / status 过滤）
func (ctl *ProductController) List(c *gin.Context) {
	var req service.ListProductRequest
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

// Update 更新商品 PUT /api/v1/admin/products/:id
func (ctl *ProductController) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req service.UpdateProductRequest
	if !bindJSON(c, &req) {
		return
	}
	p, err := ctl.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, p)
}

// UpdateStatus 上下架切换 PUT /api/v1/admin/products/:id/status
func (ctl *ProductController) UpdateStatus(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req service.UpdateProductStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	p, err := ctl.svc.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, p)
}
