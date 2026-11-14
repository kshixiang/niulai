package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置根结构，字段与 config.yaml 一一对应
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Log    LogConfig    `mapstructure:"log"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug / release
}

// MySQLConfig 数据库连接配置
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 单位：分钟
}

// RedisConfig 缓存/会话配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig 鉴权配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"` // 单位：小时
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

// GlobalConfig 进程级全局配置单例
var GlobalConfig *Config

// InitConfig 加载并解析配置。查找顺序：
// 1) ./configs  2) ../configs  3) ../../configs
// 同时支持 APP_ 前缀的环境变量覆盖（生产环境注入敏感配置）。
func InitConfig() error {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	v.SetEnvPrefix("APP")
	// 嵌套配置 mysql.host 对应 APP_MYSQL_HOST，而不是带点的 APP_MYSQL.HOST
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return err
	}

	GlobalConfig = &cfg
	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("mysql.charset", "utf8mb4")
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.max_open_conns", 100)
	v.SetDefault("mysql.conn_max_lifetime", 60)
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.expire", 24)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.path", "./logs")
}

func validate(cfg *Config) error {
	if cfg.MySQL.Host == "" || cfg.MySQL.Database == "" {
		return fmt.Errorf("mysql.host 与 mysql.database 为必填项")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret 为必填项")
	}
	return nil
}
