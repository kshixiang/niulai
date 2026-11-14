package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"niulai/internal/service"
	"niulai/pkg/response"
)

// bindJSON 绑定并校验 JSON 请求体；校验失败时直接返回 400 并中止处理链。
func bindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.Error(err) // 记录详细错误便于排查
		response.Fail(c, response.CodeInvalidParam, "请求参数校验不通过")
		return false
	}
	return true
}

// bindQuery 绑定并校验 URL 查询参数；校验失败时直接返回 400 并中止处理链。
func bindQuery(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		c.Error(err)
		response.Fail(c, response.CodeInvalidParam, "查询参数校验不通过")
		return false
	}
	return true
}

// parseIDParam 解析路由中的 :id 参数，非法时返回 400。
func parseIDParam(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, response.CodeInvalidParam, "非法的ID参数")
		return 0, false
	}
	return id, true
}

// normalizePage 标准化分页参数，避免越界或缺失。
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// handleServiceError 将 service 层错误映射为统一响应。
func handleServiceError(c *gin.Context, err error) {
	switch {
	// 登录/注册业务失败用 400，避免前端拦截器把 401 当作会话失效清 token 并跳转
	case errors.Is(err, service.ErrPhoneExists),
		errors.Is(err, service.ErrInvalidCredential),
		errors.Is(err, service.ErrPasswordTooLong):
		response.Fail(c, response.CodeInvalidParam, err.Error())
	case errors.Is(err, service.ErrUserDisabled):
		response.Fail(c, response.CodeForbidden, err.Error())
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrProductNotFound),
		errors.Is(err, service.ErrChannelNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, response.CodeNotFound, err.Error())
	case errors.Is(err, service.ErrChannelUnusable),
		errors.Is(err, service.ErrInvalidPrice):
		response.Fail(c, response.CodeInvalidParam, err.Error())
	default:
		c.Error(err)
		response.Fail(c, response.CodeError, "服务异常，请稍后重试")
	}
}
