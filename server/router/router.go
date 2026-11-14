package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"niulai/internal/controller"
	"niulai/internal/middleware"
	"niulai/internal/service"
)

// Register 注册全部路由。
// admin 分组统一经过管理员鉴权中间件；各 service / controller 依赖在此装配。
func Register(r *gin.Engine, db *gorm.DB) {
	v1 := r.Group("/api/v1")

	// 封面公开可读：<img> 不带 Authorization，必须挂在免鉴权路径下
	uploadCtl := controller.NewUploadController()
	v1.GET("/uploads/:filename", uploadCtl.ServeUpload)

	// C 端公开接口：无需鉴权，只返回可对外展示的字段
	clientSvc := service.NewClientService(db)
	clientCtl := controller.NewClientController(clientSvc)
	v1.GET("/products", clientCtl.ListProducts)
	v1.GET("/orders", clientCtl.ListOrders)
	v1.GET("/agents", clientCtl.ListAgents)

	// C 端用户认证：注册/登录免鉴权，资料查询需 UserRequired
	userSvc := service.NewUserService(db)
	userCtl := controller.NewUserController(userSvc)
	auth := v1.Group("/auth")
	auth.POST("/register", userCtl.Register)
	auth.POST("/login", userCtl.Login)
	auth.GET("/profile", middleware.UserRequired(), userCtl.Profile)

	// 登录接口独立成免鉴权组，必须在 AdminRequired 之前注册
	public := v1.Group("/admin")
	authCtl := controller.NewAuthController()
	public.POST("/login", authCtl.Login)

	admin := v1.Group("/admin")
	admin.Use(middleware.AdminRequired())
	admin.POST("/upload/cover", uploadCtl.UploadCover)

	// 渠道管理
	channelSvc := service.NewChannelService(db)
	channelCtl := controller.NewChannelController(channelSvc)
	channels := admin.Group("/channels")
	channels.POST("", channelCtl.Create)
	channels.GET("", channelCtl.List)
	channels.PUT("/:id", channelCtl.Update)

	// 商品管理
	productSvc := service.NewProductService(db)
	productCtl := controller.NewProductController(productSvc)
	products := admin.Group("/products")
	products.POST("", productCtl.Create)
	products.GET("", productCtl.List)
	products.PUT("/:id", productCtl.Update)
	products.PUT("/:id/status", productCtl.UpdateStatus)

	// 用户管理（已注册的 C 端用户）：查询与启用/封禁
	adminUserSvc := service.NewAdminUserService(db)
	adminUserCtl := controller.NewAdminUserController(adminUserSvc)
	users := admin.Group("/users")
	users.GET("", adminUserCtl.List)
	users.PUT("/:id/status", adminUserCtl.UpdateStatus)
}
