package model

import "time"

type Node struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"type:text" json:"name"`
	Address        string     `gorm:"type:text" json:"address"`
	Protocol       string     `gorm:"type:text" json:"protocol"`
	Port           int        `json:"port"`
	ConfigMode     string     `gorm:"type:text;default:auto" json:"config_mode"`
	ConfigJSON     string     `gorm:"type:text" json:"config_json"`
	Online         bool       `gorm:"default:false" json:"online"`
	LastHeartbeat  *time.Time `json:"last_heartbeat"`
	RegisterToken  string     `gorm:"type:text" json:"-"`
	Sort           int        `gorm:"default:0" json:"sort"`
	Status         int        `gorm:"default:1" json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Node) TableName() string {
	return "nodes"
}
