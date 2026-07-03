package model

import "time"

type Plan struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"type:text" json:"name"`
	Description       string    `gorm:"type:text" json:"description"`
	GroupID           *uint     `json:"group_id"`
	TrafficLimit      int64     `gorm:"default:0" json:"traffic_limit"`
	DurationDays      int       `json:"duration_days"`
	Price             int64     `gorm:"default:0" json:"price"`
	SpeedLimit        int       `gorm:"default:0" json:"speed_limit"`
	DeviceLimit       int       `gorm:"default:0" json:"device_limit"`
	CapacityLimit     int       `gorm:"default:0" json:"capacity_limit"`
	TrafficReset      int       `gorm:"default:0" json:"traffic_reset"`
	Sort              int       `gorm:"default:0" json:"sort"`
	Status            int       `gorm:"default:1" json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (Plan) TableName() string {
	return "plans"
}