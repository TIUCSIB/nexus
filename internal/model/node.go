package model

import "time"

type Node struct {
	ID              uint    `gorm:"primaryKey" json:"id"`
	CustomID        string  `gorm:"type:text" json:"custom_id"`
	Name            string  `gorm:"type:text" json:"name"`
	Address         string  `gorm:"type:text" json:"address"`
	Protocol        string  `gorm:"type:text" json:"protocol"`
	Port            int     `json:"port"`
	ServicePort     int     `gorm:"default:0" json:"service_port"`
	GroupID         *uint   `json:"group_id"`                         // 兼容旧数据
	GroupIDs        []uint  `gorm:"serializer:json" json:"group_ids"` // 多分组
	RouteID         *uint   `json:"route_id"`
	MachineID       *uint   `json:"machine_id"` // 关联机器
	Rate            float64 `gorm:"default:1" json:"rate"`
	DynamicRate     bool    `gorm:"default:false" json:"dynamic_rate"`
	Tags            string  `gorm:"type:text" json:"tags"`
	TrafficLimit    int64   `gorm:"default:0" json:"traffic_limit"`
	TrafficUsed     int64   `gorm:"default:0" json:"traffic_used"`
	OnlineCount     int     `gorm:"default:0" json:"online_count"`
	ParentID        *uint   `json:"parent_id"`
	Security        string  `gorm:"type:text;default:none" json:"security"`
	Transport       string  `gorm:"type:text;default:tcp" json:"transport"`
	FlowControl     string  `gorm:"type:text;default:none" json:"flow_control"`
	VlessEncryption bool    `gorm:"default:false" json:"vless_encryption"`
	ConfigMode      string  `gorm:"type:text;default:auto" json:"config_mode"`
	NetworkSettings string  `gorm:"type:text" json:"network_settings"`
	ConfigJSON      string  `gorm:"type:text" json:"config_json"`
	// CertConfig holds per-node TLS certificate automation config (JSON).
	// Modes: none/file/content/self/http/dns. See CertConfig struct in model/cert_config.go.
	CertConfig string `gorm:"type:text" json:"cert_config"`
	// KernelType selects the proxy kernel: "singbox" (default) or "xray".
	// Currently only "singbox" is implemented; the field is transparently
	// forwarded to the agent for future use.
	KernelType string `gorm:"type:text;default:singbox" json:"kernel_type"`
	// CustomOutbounds is a JSON array of outbound specs (tag/protocol/settings/proxy_tag)
	// pushed to the agent for落地链/链式代理. See model/custom_outbound.go.
	CustomOutbounds string     `gorm:"type:text" json:"custom_outbounds"`
	Online          bool       `gorm:"default:false" json:"online"`
	LastHeartbeat   *time.Time `json:"last_heartbeat"`
	CPU             float64    `gorm:"default:0" json:"cpu"`
	Mem             float64    `gorm:"default:0" json:"mem"`
	Uptime          uint64     `gorm:"default:0" json:"uptime"`
	RegisterToken   string     `gorm:"type:text" json:"-"`
	Sort            int        `gorm:"default:0" json:"sort"`
	Status          int        `gorm:"default:1" json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (Node) TableName() string {
	return "nodes"
}
