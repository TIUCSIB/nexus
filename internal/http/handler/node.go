package handler

import (
	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

func ListNodes(c *gin.Context) {
	var nodes []model.Node
	database.DB.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&nodes)

	if nodes == nil {
		nodes = []model.Node{}
	}

	Success(c, nodes)
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
