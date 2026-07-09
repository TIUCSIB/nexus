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

// nodeAllowedGroups returns the deduplicated set of user groups a node is
// allowed to serve, merging the legacy GroupID field with the multi-group
// GroupIDs slice. An empty result means the node has no group restriction
// (open node) and serves every active user.
func nodeAllowedGroups(node *nexusmodel.Node) []uint {
	seen := make(map[uint]struct{})
	var allowed []uint
	if node.GroupID != nil && *node.GroupID > 0 {
		seen[*node.GroupID] = struct{}{}
		allowed = append(allowed, *node.GroupID)
	}
	for _, gid := range node.GroupIDs {
		if gid == 0 {
			continue
		}
		if _, ok := seen[gid]; ok {
			continue
		}
		seen[gid] = struct{}{}
		allowed = append(allowed, gid)
	}
	return allowed
}

// GetActiveUsersForNode returns the users that the given node is currently
// allowed to serve: status is active, not expired, has not exceeded their
// traffic quota, and either the node is open (no group restriction) or the
// user's group matches one of the node's allowed groups. Users without a
// group are treated as open users and can use any node.
func GetActiveUsersForNode(node *nexusmodel.Node) ([]nexusmodel.User, error) {
	var users []nexusmodel.User
	q := nexusdb.DB.Where("status = ?", 1)
	q = q.Where("expired_at IS NULL OR expired_at > ?", time.Now())
	q = q.Where("traffic_limit = 0 OR traffic_used < traffic_limit")
	if err := q.Find(&users).Error; err != nil {
		return nil, err
	}

	allowed := nodeAllowedGroups(node)
	if len(allowed) == 0 {
		return users, nil
	}

	allowedSet := make(map[uint]struct{}, len(allowed))
	for _, gid := range allowed {
		allowedSet[gid] = struct{}{}
	}

	filtered := users[:0]
	for _, u := range users {
		if u.GroupID == nil || *u.GroupID == 0 {
			// Open user: no group restriction.
			filtered = append(filtered, u)
			continue
		}
		if _, ok := allowedSet[*u.GroupID]; ok {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
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
