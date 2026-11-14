package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"niulai/internal/config"
	"niulai/pkg/jwt"
	"niulai/pkg/response"
)

// UserRequired C 端用户鉴权：校验 Bearer JWT 的算法、签名、有效期与用户角色，
// 解析成功后把用户ID写入上下文供后续 handler 使用。
func UserRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			response.Fail(c, response.CodeUnauthorized, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		claims, err := jwt.Verify(strings.TrimPrefix(auth, "Bearer "), config.GlobalConfig.JWT.Secret)
		if err != nil || claims.Role != jwt.RoleUser {
			response.Fail(c, response.CodeUnauthorized, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(claims.Sub, 10, 64)
		if err != nil || userID == 0 {
			response.Fail(c, response.CodeUnauthorized, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
