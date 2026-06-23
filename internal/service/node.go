package service

import (
	"time"

	nexusdb "nexus/internal/database"
	nexusmodel "nexus/internal/model"
)

// GetNodeByID returns a single node by its primary key.
func GetNodeByID(id uint) (*nexusmodel.Node, error) {
	var node nexusmodel.Node
	if err := nexusdb.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

// GetNodeByToken looks up a node by its pre-shared RegisterToken.
func GetNodeByToken(token string) (*nexusmodel.Node, error) {
	var node nexusmodel.Node
	if err := nexusdb.DB.Where("register_token = ?", token).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

// UpdateNodeStatus persists the latest heartbeat metrics for a node.
func UpdateNodeStatus(nodeID uint, online bool, cpuUsage, memoryUsage float64) error {
	now := time.Now()
	return nexusdb.DB.Model(&nexusmodel.Node{}).
		Where("id = ?", nodeID).
		Updates(map[string]any{
			"online":         online,
			"last_heartbeat": &now,
		}).Error
}

// UpdateNodeInfo refreshes the name and address a node reported during
// registration and marks it online.
func UpdateNodeInfo(nodeID uint, name, address string) error {
	now := time.Now()
	return nexusdb.DB.Model(&nexusmodel.Node{}).
		Where("id = ?", nodeID).
		Updates(map[string]any{
			"name":           name,
			"address":        address,
			"online":         true,
			"last_heartbeat": &now,
		}).Error
}

// MarkNodeOffline sets a node online flag to false.
func MarkNodeOffline(nodeID uint) error {
	return nexusdb.DB.Model(&nexusmodel.Node{}).
		Where("id = ?", nodeID).
		Update("online", false).Error
}

// GetActiveUsers returns every user that is currently eligible to use the
// proxy: status is active, not expired, and has not exceeded their traffic
// quota (a zero TrafficLimit means unlimited).
func GetActiveUsers() ([]nexusmodel.User, error) {
	var users []nexusmodel.User
	q := nexusdb.DB.Where("status = ?", 1)
	q = q.Where("expired_at IS NULL OR expired_at > ?", time.Now())
	q = q.Where("traffic_limit = 0 OR traffic_used < traffic_limit")
	if err := q.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// HasNodeConfigChanged returns true when the node updated_at timestamp is
// more recent than the given lastSync time.
func HasNodeConfigChanged(nodeID uint, lastSync time.Time) (bool, error) {
	var node nexusmodel.Node
	if err := nexusdb.DB.Select("updated_at").First(&node, nodeID).Error; err != nil {
		return false, err
	}
	return node.UpdatedAt.After(lastSync), nil
}
