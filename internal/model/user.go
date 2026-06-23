package model

import "time"

type User struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UUID          string     `gorm:"type:text;uniqueIndex" json:"uuid"`
	Email         string     `gorm:"type:text;uniqueIndex" json:"email"`
	PasswordHash  string     `gorm:"type:text" json:"-"`
	Balance       int64      `gorm:"default:0" json:"balance"`
	PlanID        *uint      `json:"plan_id"`
	TrafficUsed   int64      `gorm:"default:0" json:"traffic_used"`
	TrafficLimit  int64      `gorm:"default:0" json:"traffic_limit"`
	ExpiredAt     *time.Time `json:"expired_at"`
	IsAdmin       bool       `gorm:"default:false" json:"is_admin"`
	Token         string     `gorm:"type:text;uniqueIndex" json:"token"`
	Status        int        `gorm:"default:1" json:"status"`
	DeviceLimit   int        `gorm:"default:0" json:"device_limit"`
	SpeedLimitUp  int        `gorm:"default:0" json:"speed_limit_up"`
	SpeedLimitDown int       `gorm:"default:0" json:"speed_limit_down"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
