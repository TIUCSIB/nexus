package model

import "time"

// RouteRule represents a routing rule that gets distributed to proxy nodes.
// Rules can block domains, redirect traffic, set custom DNS, etc.
type RouteRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:text" json:"name"`
	Match       string    `gorm:"type:text" json:"match"`        // legacy newline separated match list
	Action      string    `gorm:"type:text" json:"action"`       // direct/block/route/proxy
	ActionValue string    `gorm:"type:text" json:"action_value"` // outbound tag for route/proxy
	MatchJSON   string    `gorm:"type:text" json:"match_json"`   // structured RouteMatch JSON
	ActionJSON  string    `gorm:"type:text" json:"action_json"`  // structured RouteAction JSON
	Sort        int       `gorm:"default:0" json:"sort"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (RouteRule) TableName() string {
	return "route_rules"
}
