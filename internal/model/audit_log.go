package model

import "time"

// AuditLog 操作审计日志
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null;default:0" json:"user_id"`
	UserEmail string    `gorm:"size:255;not null;default:''" json:"user_email"`
	Action    string    `gorm:"size:64;index;not null;default:''" json:"action"`
	Target    string    `gorm:"size:255;not null;default:''" json:"target"`
	Detail    string    `gorm:"type:text" json:"detail"`
	IP        string    `gorm:"size:64;not null;default:''" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}