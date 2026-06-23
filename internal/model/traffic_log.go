package model

import "time"

type TrafficLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	NodeID     uint      `gorm:"index" json:"node_id"`
	Upload     int64     `json:"upload"`
	Download   int64     `json:"download"`
	RecordedAt time.Time `gorm:"index" json:"recorded_at"`
}

func (TrafficLog) TableName() string {
	return "traffic_logs"
}
