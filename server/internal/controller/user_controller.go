package controller

import (
	"github.com/gin-gonic/gin"

	"niulai/internal/config"
	"niulai/internal/service"
	"niulai/pkg/jwt"
	"niulai/pkg/response"
)

// UserController C 端用户认证接口（/api/v1/auth/*）。
type UserController struct {
	svc *service.UserService
}

func NewUserController(svc *service.UserService) *UserController {
	return &UserController{svc: svc}
}

// Register 用户注册 POST /api/v1/auth/register
func (ctl *UserController) Register(c *gin.Context) {
	var req service.UserRegisterRequest
	if !bindJSON(c, &req) {
		return
	}
	view, err := ctl.svc.Register(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	ctl.respondWithToken(c, view)
}

// Login 用户登录 POST /api/v1/auth/login
func (ctl *UserController) Login(c *gin.Context) {
	var req service.UserLoginRequest
	if !bindJSON(c, &req) {
		return
	}
	view, err := ctl.svc.Login(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	ctl.respondWithToken(c, view)
}

// Profile 当前用户资料 GET /api/v1/auth/profile
func (ctl *UserController) Profile(c *gin.Context) {
	view, err := ctl.svc.GetProfile(c.Request.Context(), c.GetUint64("user_id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, view)
}

// UpdateProfile 修改当前用户昵称 PUT /api/v1/auth/profile
func (ctl *UserController) UpdateProfile(c *gin.Context) {
	var req service.UserProfileUpdateRequest
	if !bindJSON(c, &req) {
		return
	}
	view, err := ctl.svc.UpdateProfile(c.Request.Context(), c.GetUint64("user_id"), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, view)
}

// respondWithToken 注册/登录成功后签发令牌，实现注册即登录。
func (ctl *UserController) respondWithToken(c *gin.Context, view *service.UserView) {
	cfg := config.GlobalConfig
	token, err := jwt.SignUser(view.ID, cfg.JWT.Secret, cfg.JWT.Expire)
	if err != nil {
		c.Error(err)
		response.Fail(c, response.CodeError, "操作失败，请稍后重试")
		return
	}
	response.Success(c, gin.H{"token": token, "user": view})
}
