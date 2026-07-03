package model

import "time"

// NodeAuth stores authentication tokens for agent connections.
// Each node gets a unique auth token when the agent registers.
type NodeAuth struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NodeID    uint      `gorm:"uniqueIndex" json:"node_id"`
	AuthToken string    `gorm:"type:text;uniqueIndex" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

func (NodeAuth) TableName() string {
	return "node_auths"
}
