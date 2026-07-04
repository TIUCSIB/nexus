package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---------- Request / Response types ----------

type agentHeartbeatReq struct {
	CPU    float64 `json:"cpu"`
	Mem    float64 `json:"mem"`
	Uptime uint64  `json:"uptime"`
}

type agentHeartbeatResp struct {
	ConfigChanged bool `json:"config_changed"`
}

type agentConfigResp struct {
	ConfigJSON string `json:"config_json"`
	UsersJSON  string `json:"users_json"`
	RoutesJSON string `json:"routes_json"`
}

type agentTrafficEntry struct {
	UserUUID string `json:"user_uuid" binding:"required"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

type agentAliveReq struct {
	Data map[string][]string `json:"data" binding:"required"`
}

type agentAliveResp struct {
	Alive map[string]int `json:"alive"`
}

// ---------- Middleware: server token auth ----------

// ServerAuthMiddleware validates the global server_token and injects node_id into context.
// Token is sent via X-Node-Token header, node_id is extracted from the URL path.
func ServerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Node-Token")
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "missing X-Node-Token header",
			})
			c.Abort()
			return
		}

		serverToken := database.GetSetting("server_token")
		if serverToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": "server_token not configured",
			})
			c.Abort()
			return
		}

		if token != serverToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "invalid server token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ---------- Handlers ----------

// AgentHeartbeat handles periodic agent heartbeats.
// POST /api/internal/agent/:node_id/heartbeat
func AgentHeartbeat(c *gin.Context) {
	nodeID := c.Param("node_id")

	var req agentHeartbeatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Update node status
	now := time.Now()
	result := database.DB.Model(&model.Node{}).Where("id = ?", nodeID).Updates(map[string]interface{}{
		"online":         true,
		"last_heartbeat": &now,
	})
	if result.Error != nil {
		InternalError(c, "failed to update node status")
		return
	}

	if result.RowsAffected == 0 {
		NotFound(c, "node not found")
		return
	}

	// Check if config has been updated since last heartbeat.
	var node model.Node
	configChanged := false
	if err := database.DB.First(&node, nodeID).Error; err == nil {
		if node.LastHeartbeat != nil && node.UpdatedAt.After(*node.LastHeartbeat) {
			configChanged = true
		}
	}

	Success(c, agentHeartbeatResp{
		ConfigChanged: configChanged,
	})
}

// AgentGetConfig returns the sing-box configuration for the requesting node.
// GET /api/internal/agent/:node_id/config
func AgentGetConfig(c *gin.Context) {
	nodeID := c.Param("node_id")

	var node model.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		NotFound(c, "node not found")
		return
	}

	// Fetch active users for config generation
	var users []model.User
	database.DB.Where("status = ?", 1).Find(&users)

	// Generate complete sing-box config using the config generator
	generatedConfig, err := service.GenerateSingboxConfig(node, users)
	if err != nil {
		InternalError(c, "failed to generate config: "+err.Error())
		return
	}

	// Build users JSON for agent (speed limits etc)
	usersJSON := buildUsersJSON()

	// Build routes JSON for agent
	routesJSON := buildRoutesJSON()

	Success(c, agentConfigResp{
		ConfigJSON: generatedConfig,
		UsersJSON:  usersJSON,
		RoutesJSON: routesJSON,
	})
}

// AgentReportTraffic records traffic data from the agent (delta mode).
// POST /api/internal/agent/:node_id/traffic
func AgentReportTraffic(c *gin.Context) {
	nodeID := c.Param("node_id")

	var req agentTrafficEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		// Try array format for backwards compatibility
		var entries []agentTrafficEntry
		if err2 := c.ShouldBindJSON(&entries); err2 != nil {
			BadRequest(c, "invalid request: "+err.Error())
			return
		}
		processTrafficEntries(c, nodeID, entries)
		return
	}

	processTrafficEntries(c, nodeID, []agentTrafficEntry{req})
}

func processTrafficEntries(c *gin.Context, nodeID string, entries []agentTrafficEntry) {
	now := time.Now()
	recorded := 0
	for _, entry := range entries {
		// Look up user by UUID
		var user model.User
		if err := database.DB.Where("uuid = ?", entry.UserUUID).First(&user).Error; err != nil {
			continue // skip unknown users
		}

		// Record traffic log
		trafficLog := model.TrafficLog{
			UserID:     user.ID,
			NodeID:     parseNodeID(nodeID),
			Upload:     entry.Upload,
			Download:   entry.Download,
			RecordedAt: now,
		}
		if err := database.DB.Create(&trafficLog).Error; err != nil {
			log.Printf("[agent] failed to record traffic for user %s: %v", entry.UserUUID, err)
			continue
		}

		// Add delta to user's cumulative traffic usage
		delta := entry.Upload + entry.Download
		if delta > 0 {
			database.DB.Model(&user).UpdateColumn("traffic_used",
				gorm.Expr("traffic_used + ?", delta))
		}
		recorded++
	}

	log.Printf("[agent] recorded traffic for %d users (node %s)", recorded, nodeID)
	Success(c, nil)
}

// AgentReportAlive receives online IP data from agents.
// POST /api/internal/agent/:node_id/alive
func AgentReportAlive(c *gin.Context) {
	nodeID := c.Param("node_id")

	var req agentAliveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	now := time.Now()
	processed := 0
	for userUUID, ips := range req.Data {
		// Look up user by UUID
		var user model.User
		if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
			continue // skip unknown users
		}

		// Marshal IPs to JSON
		ipsJSON, err := json.Marshal(ips)
		if err != nil {
			continue
		}

		// Upsert alive IP record for this user+node combination
		var existing model.AliveIP
		err = database.DB.Where("user_id = ? AND node_id = ?", user.ID, parseNodeID(nodeID)).First(&existing).Error
		if err == nil {
			// Update existing record
			database.DB.Model(&existing).Updates(map[string]interface{}{
				"ips":        string(ipsJSON),
				"updated_at": now,
			})
		} else {
			// Create new record
			aliveIP := model.AliveIP{
				UserID:    user.ID,
				NodeID:    parseNodeID(nodeID),
				IPs:       string(ipsJSON),
				UpdatedAt: now,
			}
			database.DB.Create(&aliveIP)
		}
		processed++
	}

	// Clean up stale records (older than 120 seconds)
	staleThreshold := now.Add(-120 * time.Second)
	database.DB.Where("updated_at < ?", staleThreshold).Delete(&model.AliveIP{})

	log.Printf("[agent] processed alive data for %d users (node %s)", processed, nodeID)
	Success(c, nil)
}

// AgentGetAliveList returns count of online IPs per user (only for users with device_limit > 0).
// GET /api/internal/agent/alivelist
func AgentGetAliveList(c *gin.Context) {
	// Find all alive records for users with device_limit > 0
	var results []struct {
		UserUUID string
		Count    int
	}

	database.DB.Raw(`
		SELECT u.uuid as user_uuid, COUNT(DISTINCT a.id) as count
		FROM alive_ips a
		INNER JOIN users u ON u.id = a.user_id
		WHERE u.device_limit > 0
		GROUP BY u.uuid
	`).Scan(&results)

	alive := make(map[string]int)
	for _, r := range results {
		alive[r.UserUUID] = r.Count
	}

	Success(c, agentAliveResp{Alive: alive})
}

// AgentGetDeviceLimit returns device_limit for all users that have it set.
// GET /api/internal/agent/devicelimit
func AgentGetDeviceLimit(c *gin.Context) {
	var results []struct {
		UUID        string
		DeviceLimit int
	}

	database.DB.Raw(`
		SELECT uuid, device_limit
		FROM users
		WHERE status = 1 AND device_limit > 0
	`).Scan(&results)

	limits := make(map[string]int)
	for _, r := range results {
		limits[r.UUID] = r.DeviceLimit
	}

	Success(c, gin.H{"limits": limits})
}

// buildUsersJSON creates a JSON array of active user credentials for sing-box.
func buildUsersJSON() string {
	var users []model.User
	database.DB.Where("status = ?", 1).Find(&users)

	result := "["
	for i, u := range users {
		if i > 0 {
			result += ","
		}
		result += `{"uuid":"` + u.UUID + `","email":"` + u.Email + `"`
		if u.SpeedLimitUp > 0 {
			result += `,"speed_limit_up":` + fmt.Sprintf("%d", u.SpeedLimitUp)
		}
		if u.SpeedLimitDown > 0 {
			result += `,"speed_limit_down":` + fmt.Sprintf("%d", u.SpeedLimitDown)
		}
		result += "}"
	}
	result += "]"
	return result
}

// buildRoutesJSON creates a JSON array of active route rules for distribution to agents.
func buildRoutesJSON() string {
	var rules []model.RouteRule
	database.DB.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&rules)

	result := "["
	for i, r := range rules {
		if i > 0 {
			result += ","
		}
		result += `{"id":` + fmt.Sprintf("%d", r.ID)
		result += `,"name":"` + r.Name + `"`
		result += `,"match":"` + r.Match + `"`
		result += `,"action":"` + r.Action + `"`
		if r.ActionValue != "" {
			result += `,"action_value":"` + r.ActionValue + `"`
		}
		result += "}"
	}
	result += "]"
	return result
}

// parseNodeID parses a node ID string to uint, returns 0 on error.
func parseNodeID(id string) uint {
	var n uint
	fmt.Sscanf(id, "%d", &n)
	return n
}
