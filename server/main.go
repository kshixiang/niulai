package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"niulai/internal/config"
	"niulai/pkg/database"
	"niulai/pkg/response"
)

func main() {
	// 1. 配置
	if err := config.InitConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}
	cfg := config.GlobalConfig

	// 2. 数据库
	db, err := database.InitDB(&cfg.MySQL)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	// 仅非生产环境自动迁移表结构，生产环境请移除该行
	if cfg.Server.Mode != "release" {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("自动迁移表结构失败: %v", err)
		}
	}

	// 3. 路由
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// 4. 启动 HTTP 服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}
	go func() {
		log.Printf("服务启动在 :%d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 5. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭异常: %v", err)
	}
	log.Println("服务已退出")
}
