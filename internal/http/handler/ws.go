package handler

import (
	"fmt"
	"net/http"

	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/ws"

	"github.com/gin-gonic/gin"
)

// WSHub is the global WebSocket hub for agent connections.
var WSHub = ws.NewHub()

// AgentWebSocket handles the WebSocket upgrade for agent connections.
// Supports two auth modes:
//   - node mode: ?node_id=N&token=<server_token>
//   - machine mode: ?machine_id=N&token=<machine_token>
//
// GET /api/internal/agent/ws
func AgentWebSocket(c *gin.Context) {
	nodeID := c.Query("node_id")
	machineID := c.Query("machine_id")
	token := c.Query("token")

	if token == "" {
		token = c.GetHeader("X-Node-Token")
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "missing authentication"})
		return
	}

	// Machine mode auth
	if machineID != "" {
		var machine model.Machine
		if err := database.DB.Where("id = ?", machineID).First(&machine).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "machine not found"})
			return
		}
		if token != machine.Token {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "invalid machine token"})
			return
		}
		if !machine.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "machine is disabled"})
			return
		}

		// Load all active nodes under this machine
		var nodes []model.Node
		database.DB.Where("machine_id = ? AND status = 1", machine.ID).Find(&nodes)
		nodeIDs := make([]string, len(nodes))
		for i, n := range nodes {
			// Use string representation of the node ID
			nodeIDs[i] = fmt.Sprintf("%d", n.ID)
		}

		// Upgrade with machine identity and pre-populated node list
		WSHub.ServeWS(c.Writer, c.Request, "", fmt.Sprintf("%d", machine.ID), token)
		// Register nodes after connection is established
		WSHub.RegisterMachineNodes(fmt.Sprintf("%d", machine.ID), nodeIDs)
		return
	}

	// Node mode auth (legacy)
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "missing node_id or machine_id"})
		return
	}

	serverToken := database.GetSetting("server_token")
	if serverToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "server_token not configured"})
		return
	}

	if token != serverToken {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "invalid server token"})
		return
	}

	// Upgrade to WebSocket (node mode)
	WSHub.ServeWS(c.Writer, c.Request, nodeID, "", token)
}

// AdminSendNodeCommand sends a command to a connected agent via WebSocket.
// POST /api/admin/nodes/:id/command
func AdminSendNodeCommand(c *gin.Context) {
	nodeID := c.Param("id")

	var req struct {
		Command string `json:"command" binding:"required"`
		Data    string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请填写指令")
		return
	}

	cmd := &ws.Command{
		Type: req.Command,
	}
	if req.Data != "" {
		cmd.Data = []byte(req.Data)
	}

	if err := WSHub.SendCommand(nodeID, cmd); err != nil {
		BadRequest(c, "节点未连接 WebSocket")
		return
	}

	Success(c, gin.H{"message": "指令已发送", "node_id": nodeID})
}