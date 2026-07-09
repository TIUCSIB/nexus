package model

import "time"

// LoadStatus represents real-time system load data reported by the agent.
type LoadStatus struct {
	CPU      float64 `json:"cpu"`
	MemTotal int64   `json:"mem_total"`
	MemUsed  int64   `json:"mem_used"`
	DiskTotal int64  `json:"disk_total"`
	DiskUsed  int64  `json:"disk_used"`
	NetInSpeed  float64 `json:"net_in_speed"`
	NetOutSpeed float64 `json:"net_out_speed"`
}

// Machine represents a physical machine that hosts one or more proxy nodes.
// Reference: Xboard App\Models\ServerMachine
type Machine struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"type:text" json:"name"`
	Token      string     `gorm:"type:text" json:"-"`
	Notes      string     `gorm:"type:text" json:"notes,omitempty"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	LoadStatus *LoadStatus `gorm:"serializer:json" json:"load_status,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (Machine) TableName() string {
	return "machines"
}

// ServerCount is a view-only field populated by the admin handler, not stored in DB.
type MachineWithCount struct {
	Machine
	ServerCount int64 `json:"servers_count"`
}