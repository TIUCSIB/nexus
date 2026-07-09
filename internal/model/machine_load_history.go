package model

import "time"

// MachineLoadHistory stores historical system load data for a machine.
// Reference: Xboard App\Models\ServerMachineLoadHistory
type MachineLoadHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	MachineID   uint      `gorm:"index;not null" json:"machine_id"`
	CPU         float64   `json:"cpu"`
	MemTotal    int64     `json:"mem_total"`
	MemUsed     int64     `json:"mem_used"`
	DiskTotal   int64     `json:"disk_total"`
	DiskUsed    int64     `json:"disk_used"`
	NetInSpeed  float64   `json:"net_in_speed"`
	NetOutSpeed float64   `json:"net_out_speed"`
	RecordedAt  int64     `json:"recorded_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (MachineLoadHistory) TableName() string {
	return "machine_load_histories"
}