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

	// 系统设置与注册码：注册配置公开只读，后台读写走 admin 鉴权组
	settingSvc := service.NewSettingService(db)
	settingCtl := controller.NewSettingController(settingSvc)
	v1.GET("/register-config", settingCtl.RegisterConfig)
	codeSvc := service.NewRegisterCodeService(db)
	codeCtl := controller.NewRegisterCodeController(codeSvc)

	// C 端公开接口：无需鉴权，只返回可对外展示的字段
	clientSvc := service.NewClientService(db)
	clientCtl := controller.NewClientController(clientSvc)
	v1.GET("/products", clientCtl.ListProducts)
	v1.GET("/products/:id", clientCtl.GetProduct)
	v1.GET("/orders/counts", clientCtl.OrderCounts)
	v1.GET("/orders", clientCtl.ListOrders)
	v1.GET("/orders/:id", clientCtl.GetOrder)
	v1.GET("/agents", clientCtl.ListAgents)

	// C 端下单：需用户登录，金额由服务端按商品折扣重算
	orderSvc := service.NewOrderService(db)
	orderCtl := controller.NewOrderController(orderSvc)
	v1.POST("/orders", middleware.UserRequired(), orderCtl.Create)

	// C 端用户认证：注册/登录免鉴权，资料查询需 UserRequired
	userSvc := service.NewUserService(db, settingSvc, codeSvc)
	userCtl := controller.NewUserController(userSvc)
	auth := v1.Group("/auth")
	auth.POST("/register", userCtl.Register)
	auth.POST("/login", userCtl.Login)
	auth.GET("/profile", middleware.UserRequired(), userCtl.Profile)
	auth.PUT("/profile", middleware.UserRequired(), userCtl.UpdateProfile)

	// C 端资金账户：需用户登录，只返回当前用户自己的余额
	walletSvc := service.NewWalletService(db)
	walletCtl := controller.NewWalletController(walletSvc)
	v1.GET("/wallet", middleware.UserRequired(), walletCtl.Get)

	// 登录接口独立成免鉴权组，必须在 AdminRequired 之前注册
	public := v1.Group("/admin")
	authCtl := controller.NewAuthController()
	public.POST("/login", authCtl.Login)

	admin := v1.Group("/admin")
	admin.Use(middleware.AdminRequired())
	admin.POST("/upload/cover", uploadCtl.UploadCover)
	admin.GET("/settings", settingCtl.Get)
	admin.PUT("/settings", settingCtl.Update)

	// 注册码管理（后台生成、启用/禁用）
	codes := admin.Group("/register-codes")
	codes.POST("", codeCtl.Generate)
	codes.GET("", codeCtl.List)
	codes.PUT("/:id/status", codeCtl.UpdateStatus)

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

	// 订单管理（列表、详情与编辑，含成本/利润等内部数据）
	adminOrderSvc := service.NewAdminOrderService(db)
	adminOrderCtl := controller.NewAdminOrderController(adminOrderSvc)
	adminOrders := admin.Group("/orders")
	adminOrders.GET("", adminOrderCtl.List)
	adminOrders.GET("/:id", adminOrderCtl.Get)
	adminOrders.PUT("/:id", adminOrderCtl.Update)
}
