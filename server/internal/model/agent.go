package model

// Agent 代理关系（多级分销）
type Agent struct {
	BaseModel
	UserID     uint64 `gorm:"column:user_id;type:bigint;not null;uniqueIndex;comment:用户ID(关联users.id)" json:"user_id"`
	ParentID   uint64 `gorm:"column:parent_id;type:bigint;not null;default:0;index;comment:上级代理ID，顶级为0" json:"parent_id"`
	Path       string `gorm:"column:path;type:varchar(255);not null;default:'0';index;comment:层级路径 如 0/1/5" json:"path"`
	Level      int8   `gorm:"column:level;type:tinyint;not null;default:1;index;comment:代理等级" json:"level"`
	InviteCode string `gorm:"column:invite_code;type:varchar(32);not null;default:'';uniqueIndex;comment:邀请码" json:"invite_code"`
}

func (Agent) TableName() string { return "agents" }
