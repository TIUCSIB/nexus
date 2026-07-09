package model

import "time"

// TrafficResetLog 流量重置日志
type TrafficResetLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null;default:0" json:"user_id"`
	UserEmail string    `gorm:"size:255;not null;default:''" json:"user_email"`
	PlanID    *uint     `json:"plan_id"`
	PlanName  string    `gorm:"size:255;not null;default:''" json:"plan_name"`
	Method    string    `gorm:"size:32;index;not null;default:''" json:"method"`
	Operator  string    `gorm:"size:255;not null;default:''" json:"operator"`
	CreatedAt time.Time `json:"created_at"`
}