package model

import "time"

// ServerGroup represents a permission group for nodes.
// Plans are bound to server groups, and users can only access nodes in their plan's group.
type ServerGroup struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:text;uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ServerGroup) TableName() string {
	return "server_groups"
}
