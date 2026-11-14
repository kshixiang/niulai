package model

// 渠道状态
const (
	ChannelStatusOffline int8 = 0
	ChannelStatusOnline  int8 = 1
)

// Channel 充值渠道/供货商。
// 敏感字段（APIEndpoint/AppKey/AppSecret）json 标签置空，避免序列化泄露。
type Channel struct {
	BaseModel
	Name        string `gorm:"column:name;type:varchar(64);not null;comment:渠道名称" json:"name"`
	Type        string `gorm:"column:type;type:varchar(32);not null;comment:渠道类型" json:"type"`
	APIEndpoint string `gorm:"column:api_endpoint;type:varchar(255);default:'';comment:接口地址" json:"-"`
	AppKey      string `gorm:"column:app_key;type:varchar(128);default:'';comment:渠道应用标识" json:"-"`
	AppSecret   string `gorm:"column:app_secret;type:varchar(255);default:'';comment:渠道密钥" json:"-"`
	Weight      int    `gorm:"column:weight;type:int;default:100;not null;comment:路由权重" json:"weight"`
	Status      int8   `gorm:"column:status;type:tinyint;default:1;not null;index;comment:状态 0-停用 1-启用" json:"status"`
}

func (Channel) TableName() string { return "channels" }
