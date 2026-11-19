package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/service"
	"niulai/pkg/response"
)

type CategoryController struct{ svc *service.CategoryService }

func NewCategoryController(svc *service.CategoryService) *CategoryController {
	return &CategoryController{svc: svc}
}

// 分类名称由 service 统一规范化与校验。
type categoryRequest struct {
	Name string `json:"name"`
}

func (ctl *CategoryController) List(c *gin.Context) {
	list, err := ctl.svc.List(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, list)
}

func (ctl *CategoryController) Create(c *gin.Context) {
	var req categoryRequest
	if !bindJSON(c, &req) {
		return
	}
	category, err := ctl.svc.Create(c.Request.Context(), req.Name)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, category)
}

func (ctl *CategoryController) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req categoryRequest
	if !bindJSON(c, &req) {
		return
	}
	category, err := ctl.svc.Update(c.Request.Context(), id, req.Name)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, category)
}
