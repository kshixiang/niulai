package middleware

import (
	"github.com/gin-gonic/gin"

	"niulai/pkg/response"
)

// AdminRequired 管理员权限拦截中间件（占位实现）。
//
// 当前仅校验请求是否携带 Authorization 头作为管理员凭证占位；
// 后续接入真实鉴权时，应在此解析 JWT、校验管理员角色声明，
// 并将管理员信息写入上下文供后续 handler 使用。
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			response.Fail(c, response.CodeUnauthorized, "缺少管理员凭证")
			c.Abort()
			return
		}
		c.Next()
	}
}
