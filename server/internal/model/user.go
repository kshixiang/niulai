package model

import "time"

// 用户状态
const (
	UserStatusDisabled int8 = 0
	UserStatusNormal   int8 = 1
)

// User 系统用户（C 端充值用户）
type User struct {
	BaseModel
	Username    string     `gorm:"column:username;type:varchar(64);uniqueIndex;not null;comment:用户名" json:"username"`
	Phone       string     `gorm:"column:phone;type:varchar(20);uniqueIndex;not null;comment:手机号" json:"phone"`
	Password    string     `gorm:"column:password;type:varchar(128);not null;comment:登录密码(BCrypt哈希)" json:"-"`
	Nickname    string     `gorm:"column:nickname;type:varchar(64);default:'';comment:昵称" json:"nickname"`
	Avatar      string     `gorm:"column:avatar;type:varchar(255);default:'';comment:头像URL" json:"avatar"`
	Status      int8       `gorm:"column:status;type:tinyint;default:1;not null;index;comment:状态 0-禁用 1-正常" json:"status"`
	LastLoginAt *time.Time `gorm:"column:last_login_at;comment:最后登录时间" json:"last_login_at"`
}

func (User) TableName() string { return "users" }
