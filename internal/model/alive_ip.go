package model

import "time"

// AliveIP stores the list of IPs currently connected for each user.
// Used for device limit enforcement. Records are created/updated by agents
// and expire after a configurable TTL.
type AliveIP struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	IPs       string    `gorm:"type:text" json:"ips"`
	NodeID    uint      `gorm:"index" json:"node_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AliveIP) TableName() string {
	return "alive_ips"
}
