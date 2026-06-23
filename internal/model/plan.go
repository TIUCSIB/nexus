package model

import "time"

type Plan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:text" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`
	TrafficLimit int64     `gorm:"default:0" json:"traffic_limit"`
	DurationDays int       `json:"duration_days"`
	Price        int64     `gorm:"default:0" json:"price"`
	Sort         int       `gorm:"default:0" json:"sort"`
	Status       int       `gorm:"default:1" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Plan) TableName() string {
	return "plans"
}
