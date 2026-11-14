package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"niulai/internal/config"
	"niulai/pkg/jwt"
	"niulai/pkg/response"
)

// AdminRequired 管理员鉴权：校验 Bearer JWT 的算法、签名与有效期。
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			response.Fail(c, response.CodeUnauthorized, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		claims, err := jwt.Verify(strings.TrimPrefix(auth, "Bearer "), config.GlobalConfig.JWT.Secret)
		// 用户令牌与管理端共用密钥，必须校验角色，防止用户令牌越权访问后台接口
		if err != nil || claims.Role != jwt.RoleAdmin {
			response.Fail(c, response.CodeUnauthorized, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set("admin_username", claims.Sub)
		c.Next()
	}
}
