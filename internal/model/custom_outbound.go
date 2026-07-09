package model

import "time"

// CustomOutbound defines a reusable outbound (落地/链式代理) that can be
// bound to nodes. The settings_json holds protocol-specific fields as a
// JSON object (e.g. server/port/uuid/...), forwarded verbatim to the
// kernel by the agent.
type CustomOutbound struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:text" json:"name"`
	Tag          string    `gorm:"type:text;uniqueIndex" json:"tag"` // unique outbound tag
	Protocol     string    `gorm:"type:text" json:"protocol"`        // vless/vmess/trojan/shadowsocks/wireguard/http/socks/...
	SettingsJSON string    `gorm:"type:text" json:"settings_json"`   // protocol-specific settings
	ProxyTag     string    `gorm:"type:text" json:"proxy_tag"`       // chain proxy: next outbound tag
	Sort         int       `gorm:"default:0" json:"sort"`
	Status       int       `gorm:"default:1" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (CustomOutbound) TableName() string { return "custom_outbounds" }

// NodeOutbound is the join table binding outbounds to nodes.
type NodeOutbound struct {
	ID               uint `gorm:"primaryKey" json:"id"`
	NodeID           uint `gorm:"index" json:"node_id"`
	CustomOutboundID uint `gorm:"index" json:"custom_outbound_id"`
}

func (NodeOutbound) TableName() string { return "node_outbounds" }
