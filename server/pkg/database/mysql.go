package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"niulai/internal/config"
	"niulai/internal/model"
)

// InitDB 连接 MySQL 并完成连接池配置，返回 *gorm.DB。
// 连接池参数来源于配置，避免高并发下连接耗尽或空闲过多。
func InitDB(cfg *config.MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // 表名使用复数形式（users/agents/...）
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 sql.DB 失败: %w", err)
	}

	// 连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL 探活失败: %w", err)
	}
	return db, nil
}

// AutoMigrate 自动迁移所有业务模型。
// 仅建议在开发/测试环境调用；生产环境请使用受版本控制的 SQL 迁移脚本。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Agent{},
		&model.Wallet{},
		&model.Product{},
		&model.ProductDenomination{},
		&model.Channel{},
		&model.Order{},
		&model.CommissionRecord{},
	)
}
