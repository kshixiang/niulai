package controller

import (
	"crypto/subtle"

	"github.com/gin-gonic/gin"

	"niulai/internal/config"
	"niulai/pkg/jwt"
	"niulai/pkg/response"
)

// LoginRequest 管理员登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthController 管理员登录接口
type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// Login 管理员登录 POST /api/v1/admin/login
func (ctl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if !bindJSON(c, &req) {
		return
	}

	cfg := config.GlobalConfig
	// 无论账号是否命中都执行两次常量时间比较，避免账号枚举与时序泄露
	userOK := subtle.ConstantTimeCompare([]byte(req.Username), []byte(cfg.Admin.Username))
	passOK := subtle.ConstantTimeCompare([]byte(req.Password), []byte(cfg.Admin.Password))
	if userOK != 1 || passOK != 1 {
		response.Fail(c, response.CodeUnauthorized, "账号或密码错误")
		return
	}

	token, err := jwt.Sign(req.Username, cfg.JWT.Secret, cfg.JWT.Expire)
	if err != nil {
		c.Error(err)
		response.Fail(c, response.CodeError, "登录失败，请稍后重试")
		return
	}
	response.Success(c, gin.H{"token": token, "username": req.Username})
}
