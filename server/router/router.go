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
	admin := v1.Group("/admin")
	admin.Use(middleware.AdminRequired())

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
}
