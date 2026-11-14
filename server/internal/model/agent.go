package model

// 代理状态
const (
	AgentStatusDisabled int8 = 0
	AgentStatusNormal   int8 = 1
)

// Agent 分销商/代理商（多级分销的核心实体）
type Agent struct {
	BaseModel
	AgentNo   string `gorm:"column:agent_no;type:varchar(32);uniqueIndex;not null;comment:代理编号" json:"agent_no"`
	UserID    uint64 `gorm:"column:user_id;type:bigint;not null;index;comment:关联用户ID" json:"user_id"`
	ParentID  uint64 `gorm:"column:parent_id;type:bigint;default:0;not null;index;comment:上级代理ID 0表示顶级" json:"parent_id"`
	Level     int8   `gorm:"column:level;type:tinyint;default:1;not null;comment:代理层级 1-3" json:"level"`
	AgentName string `gorm:"column:agent_name;type:varchar(64);not null;comment:代理名称" json:"agent_name"`
	Phone     string `gorm:"column:phone;type:varchar(20);default:'';comment:联系电话" json:"phone"`
	Rate      Money  `gorm:"column:rate;type:decimal(8,4);default:0;not null;comment:默认佣金比例" json:"rate"`
	Status    int8   `gorm:"column:status;type:tinyint;default:1;not null;index;comment:状态 0-禁用 1-正常" json:"status"`
}

func (Agent) TableName() string { return "agents" }
