package model

// 渠道状态
const (
	ChannelStatusOffline int8 = 0
	ChannelStatusOnline  int8 = 1
)

// Channel 上游渠道。接口密钥不参与 JSON 序列化，避免泄露。
type Channel struct {
	BaseModel
	Name   string `gorm:"column:name;type:varchar(128);not null;default:'';comment:渠道名称" json:"name"`
	APIURL string `gorm:"column:api_url;type:varchar(512);not null;default:'';comment:接口地址" json:"api_url"`
	APIKey string `gorm:"column:api_key;type:varchar(255);not null;default:'';comment:接口密钥" json:"-"`
	Status int8   `gorm:"column:status;type:tinyint;not null;default:1;index;comment:健康状态 1可用 0熔断" json:"status"`
}

func (Channel) TableName() string { return "channels" }
