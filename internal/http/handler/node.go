package handler

import (
	"encoding/json"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

func ListNodes(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		NotFound(c, "用户不存在")
		return
	}

	// 按用户权限组过滤节点
	query := database.DB.Where("status = ?", 1)
	if user.GroupID != nil && *user.GroupID > 0 {
		query = query.Where("(group_id = ? OR group_ids LIKE ?)", *user.GroupID, "%"+jsonString(*user.GroupID)+"%")
	}

	var nodes []model.Node
	query.Order("sort ASC, id ASC").Find(&nodes)

	// 缩减返回字段，隐藏敏感信息
	type safeNode struct {
		ID              uint   `json:"id"`
		Name            string `json:"name"`
		Address         string `json:"address"`
		Protocol        string `json:"protocol"`
		Port            int    `json:"port"`
		ServicePort     int    `json:"service_port"`
		Rate            float64 `json:"rate"`
		Tags            string `json:"tags"`
		Online          bool   `json:"online"`
		LastHeartbeat   *string `json:"last_heartbeat"`
		OnlineCount     int    `json:"online_count"`
		TrafficUsed     int64  `json:"traffic_used"`
		TrafficLimit    int64  `json:"traffic_limit"`
		Sort            int    `json:"sort"`
		Status          int    `json:"status"`
	}

	result := make([]safeNode, 0, len(nodes))
	for _, n := range nodes {
		var lastHb *string
		if n.LastHeartbeat != nil {
			t := n.LastHeartbeat.Format("2006-01-02 15:04:05")
			lastHb = &t
		}
		result = append(result, safeNode{
			ID:            n.ID,
			Name:          n.Name,
			Address:       n.Address,
			Protocol:      n.Protocol,
			Port:          n.Port,
			ServicePort:   n.ServicePort,
			Rate:          n.Rate,
			Tags:          n.Tags,
			Online:        n.Online,
			LastHeartbeat: lastHb,
			OnlineCount:   n.OnlineCount,
			TrafficUsed:   n.TrafficUsed,
			TrafficLimit:  n.TrafficLimit,
			Sort:          n.Sort,
			Status:        n.Status,
		})
	}

	Success(c, result)
}

// jsonString converts a number to its string representation for LIKE query
func jsonString(n uint) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func NodeStatus(c *gin.Context) {
	id := c.Param("id")

	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	if node.Status != 1 {
		Forbidden(c, "节点未启用")
		return
	}

	var traffic model.TrafficLog
	database.DB.Where("node_id = ?", node.ID).
		Order("recorded_at DESC").
		First(&traffic)

	Success(c, gin.H{
		"node":           node,
		"online":         node.Online,
		"last_heartbeat": node.LastHeartbeat,
		"last_traffic":   traffic,
	})
}
