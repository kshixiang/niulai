package model

// 用户状态
const (
	UserStatusDisabled int8 = 0
	UserStatusNormal   int8 = 1
)

// User 系统用户（C 端充值用户）
type User struct {
	BaseModel
	Phone    string `gorm:"column:phone;type:varchar(20);uniqueIndex;not null;default:'';comment:登录账号(手机号或自定义账号)" json:"phone"`
	Password string `gorm:"column:password;type:varchar(255);not null;default:'';comment:加密密码" json:"-"`
	Nickname string `gorm:"column:nickname;type:varchar(64);not null;default:'';comment:昵称" json:"nickname"`
	Status   int8   `gorm:"column:status;type:tinyint;not null;default:1;index;comment:状态 1正常 0封禁" json:"status"`
}

func (User) TableName() string { return "users" }
