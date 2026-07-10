package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

// ---------- Request / Response types ----------

type agentHeartbeatReq struct {
	CPU    float64 `json:"cpu"`
	Mem    float64 `json:"mem"`
	Uptime uint64  `json:"uptime,omitempty"`
	// Xboard-style fields (object format) — handled via custom unmarshaling if needed
	SwapInfo *systemInfo `json:"swap,omitempty"`
	DiskInfo *systemInfo `json:"disk,omitempty"`
}

type systemInfo struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

type agentHeartbeatResp struct {
	ConfigChanged bool `json:"config_changed"`
	PullInterval  int  `json:"pull_interval"`
}

type agentConfigResp struct {
	ConfigJSON string `json:"config_json"`
	UsersJSON  string `json:"users_json"`
	RoutesJSON string `json:"routes_json"`
}

// ---------- Xboard-style response types ----------

type handshakeResponse struct {
	WebSocket handshakeWS       `json:"websocket"`
	Settings  handshakeSettings `json:"settings"`
}

type handshakeWS struct {
	Enabled bool   `json:"enabled"`
	WSURL   string `json:"ws_url,omitempty"`
}

type handshakeSettings struct {
	PushInterval int `json:"push_interval"`
	PullInterval int `json:"pull_interval"`
}

type nodeConfigResponse struct {
	Protocol          string                 `json:"protocol"`
	ListenIP          string                 `json:"listen_ip"`
	ServerPort        int                    `json:"server_port"`
	Network           string                 `json:"network"`
	NetworkSettings   map[string]interface{} `json:"networkSettings,omitempty"`
	BaseConfig        baseConfigResp         `json:"base_config"`
	Routes            []routeRuleResp        `json:"routes"`
	KernelType        string                 `json:"kernel_type,omitempty"`
	CertConfig        map[string]interface{} `json:"cert_config,omitempty"`
	CustomOutbounds   []customOutboundResp   `json:"custom_outbounds,omitempty"`
	TLS               int                    `json:"tls,omitempty"`
	Flow              string                 `json:"flow,omitempty"`
	TLSSettings       map[string]interface{} `json:"tls_settings,omitempty"`
	ServerName        string                 `json:"server_name,omitempty"`
	UpMbps            int                    `json:"up_mbps,omitempty"`
	DownMbps          int                    `json:"down_mbps,omitempty"`
	ObfsPassword      string                 `json:"obfs-password,omitempty"`
	CongestionControl string                 `json:"congestion_control,omitempty"`
}

type baseConfigResp struct {
	PushInterval int `json:"push_interval"`
	PullInterval int `json:"pull_interval"`
}

type routeRuleResp struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Match       []string               `json:"match"`
	MatchRule   map[string]interface{} `json:"match_rule,omitempty"`
	Action      string                 `json:"action"`
	ActionValue string                 `json:"action_value,omitempty"`
	ActionRule  map[string]interface{} `json:"action_rule,omitempty"`
}

type customOutboundResp struct {
	Tag      string                 `json:"tag"`
	Protocol string                 `json:"protocol"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	ProxyTag string                 `json:"proxy_tag,omitempty"`
}

type userResp struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	SpeedLimit  int    `json:"speed_limit"`
	DeviceLimit int    `json:"device_limit"`
}

type usersResponse struct {
	Users []userResp `json:"users"`
}

type agentTrafficEntry struct {
	UserUUID string `json:"user_uuid" binding:"required"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// Xboard-style traffic format: {"user_id": [upload, download]}
type agentTrafficXboard struct {
	Data map[string][2]int64 `json:"data" binding:"required"`
}

type agentAliveReq struct {
	Data map[string][]string `json:"data" binding:"required"`
}

type agentAliveResp struct {
	Alive map[string]int `json:"alive"`
}

type agentReportReq struct {
	Traffic map[string][2]int64    `json:"traffic"`
	Alive   map[string][]string    `json:"alive"`
	Online  map[string]int         `json:"online"`
	Status  agentHeartbeatReq      `json:"status"`
	Metrics map[string]interface{} `json:"metrics"`
}

func SuccessWithETag(c *gin.Context, data interface{}) {
	b, _ := json.Marshal(data)
	sum := sha256.Sum256(b)
	etag := `"` + hex.EncodeToString(sum[:]) + `"`
	c.Header("ETag", etag)
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}
	Success(c, data)
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

// MachineAuthMiddleware validates the machine token against the Machine model.
// Token is sent via X-Machine-Token header, machine_id is extracted from the URL path.
func MachineAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Machine-Token")
		if token == "" {
			token = c.Query("token")
			if token == "" {
				token = c.Query("machine_token")
			}
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "missing machine token"})
			c.Abort()
			return
		}

		var machine model.Machine
		if err := database.DB.Where("token = ?", token).First(&machine).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "invalid machine token"})
			c.Abort()
			return
		}

		if !machine.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "machine is disabled"})
			c.Abort()
			return
		}

		c.Set("machine", &machine)
		c.Next()
	}
}

// ---------- Handlers ----------

// AgentHandshake handles the initial handshake from the agent.
// POST /api/internal/agent/handshake
func AgentHandshake(c *gin.Context) {
	pushInterval := database.GetSettingInt("node_push_interval", 30)
	pullInterval := database.GetSettingInt("node_pull_interval", 60)

	// Xboard-style WebSocket settings
	wsEnabled := database.GetSettingBool("websocket_enabled", false)

	resp := handshakeResponse{
		WebSocket: handshakeWS{
			Enabled: wsEnabled,
		},
		Settings: handshakeSettings{
			PushInterval: pushInterval,
			PullInterval: pullInterval,
		},
	}

	if wsEnabled {
		// Use custom WebSocket URL if configured, otherwise auto-detect
		wsURL := database.GetSetting("websocket_url")
		if wsURL == "" {
			scheme := "ws"
			if c.Request.TLS != nil {
				scheme = "wss"
			}
			wsURL = fmt.Sprintf("%s://%s/api/internal/agent/ws", scheme, c.Request.Host)
		}
		resp.WebSocket.WSURL = wsURL
	}

	Success(c, resp)
}

// AgentGetUsers returns the list of active users for the requesting node.
// GET /api/internal/agent/:node_id/users
func AgentGetUsers(c *gin.Context) {
	node := resolveNode(c.Param("node_id"))
	if node == nil {
		NotFound(c, "node not found")
		return
	}

	users, err := service.GetActiveUsersForNode(node)
	if err != nil {
		InternalError(c, "failed to load users: "+err.Error())
		return
	}

	result := make([]userResp, 0, len(users))
	for _, u := range users {
		// Calculate speed_limit as max(up, down) in Mbps
		speedLimit := u.SpeedLimitUp
		if u.SpeedLimitDown > speedLimit {
			speedLimit = u.SpeedLimitDown
		}

		result = append(result, userResp{
			ID:          int(u.ID),
			UUID:        u.UUID,
			SpeedLimit:  speedLimit,
			DeviceLimit: u.DeviceLimit,
		})
	}

	SuccessWithETag(c, usersResponse{Users: result})
}

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
	node := resolveNode(nodeID)
	if node == nil {
		NotFound(c, "node not found")
		return
	}
	result := database.DB.Model(&model.Node{}).Where("id = ?", node.ID).Updates(map[string]interface{}{
		"online":         true,
		"last_heartbeat": &now,
		"cpu":            req.CPU,
		"mem":            req.Mem,
		"uptime":         req.Uptime,
	})
	if result.Error != nil {
		InternalError(c, "failed to update node status")
		return
	}

	// Check if config has been updated since last heartbeat.
	configChanged := false
	if node.LastHeartbeat != nil && node.UpdatedAt.After(*node.LastHeartbeat) {
		configChanged = true
	}

	Success(c, agentHeartbeatResp{
		ConfigChanged: configChanged,
		PullInterval:  database.GetSettingInt("node_pull_interval", 60),
	})
}

// AgentGetConfig returns the node configuration for the requesting node (Xboard-style).
// GET /api/internal/agent/:node_id/config
// Supports two modes based on config_mode:
//   - "auto" (default): returns structured parameters, agent generates sing-box config
//   - "json": returns the raw config_json as a string, agent writes it directly
func AgentGetConfig(c *gin.Context) {
	nodeID := c.Param("node_id")

	var node model.Node
	if err := database.DB.Where("custom_id = ?", nodeID).First(&node).Error; err != nil {
		if err := database.DB.First(&node, nodeID).Error; err != nil {
			NotFound(c, "node not found")
			return
		}
	}

	// Mode: json — return raw config_json verbatim
	if node.ConfigMode == "json" && node.ConfigJSON != "" {
		SuccessWithETag(c, gin.H{
			"config_mode": "json",
			"config_json": node.ConfigJSON,
		})
		return
	}

	// Mode: auto (default) — return structured parameters
	pushInterval := database.GetSettingInt("node_push_interval", 30)
	pullInterval := database.GetSettingInt("node_pull_interval", 60)

	// Parse network_settings for protocol-specific parameters
	netSettings := parseJSONObject(node.NetworkSettings)
	configJSONMap := parseJSONObject(node.ConfigJSON)

	// Determine TLS mode
	tls := 0
	if node.Security == "tls" || node.Security == "reality" {
		tls = 1
	}

	// Build routes, cert config, custom outbounds
	routes := buildRoutesForNode(node)
	certConfig := parseJSONObject(node.CertConfig)
	customOutbounds := buildCustomOutboundsForNode(node)

	// Listen IP
	listenIP := "0.0.0.0"
	if node.Address != "" && node.Address != "auto" {
		listenIP = node.Address
	}

	// Network from transport
	network := node.Transport
	if network == "" {
		network = "tcp"
	}

	// Extract protocol-specific fields from network_settings (with config_json fallback)
	serverName := getStringField(netSettings, "server_name", "tls_server_name", "reality_server_name")
	if serverName == "" {
		serverName = getStringField(configJSONMap, "server_name", "handshake_server")
	}
	upMbps := getIntField(netSettings, "bandwidth_up", "up_mbps")
	if upMbps == 0 {
		upMbps = getIntField(configJSONMap, "up_mbps")
	}
	downMbps := getIntField(netSettings, "bandwidth_down", "down_mbps")
	if downMbps == 0 {
		downMbps = getIntField(configJSONMap, "down_mbps")
	}
	obfsPassword := getStringField(netSettings, "obfs_password", "obfs-password")
	if obfsPassword == "" {
		obfsPassword = getStringField(configJSONMap, "obfs_password")
	}
	congestionControl := getStringField(netSettings, "congestion_control")
	if congestionControl == "" {
		congestionControl = getStringField(configJSONMap, "congestion_control")
	}

	// Extract TLS settings from network_settings with config_json fallback
	tlsSettings := extractTLSSettings(netSettings, node.Security, node.FlowControl, configJSONMap)

	resp := nodeConfigResponse{
		Protocol:          node.Protocol,
		ListenIP:          listenIP,
		ServerPort:        node.Port,
		Network:           network,
		NetworkSettings:   netSettings,
		BaseConfig: baseConfigResp{
			PushInterval: pushInterval,
			PullInterval: pullInterval,
		},
		Routes:            routes,
		KernelType:        defaultKernelType(node.KernelType),
		CustomOutbounds:   customOutbounds,
		TLS:               tls,
		Flow:              node.FlowControl,
		ServerName:        serverName,
		UpMbps:            upMbps,
		DownMbps:          downMbps,
		ObfsPassword:      obfsPassword,
		CongestionControl: congestionControl,
	}

	// Add cert_config if available
	if len(certConfig) > 0 {
		resp.CertConfig = certConfig
	}

	// Add TLS settings if available
	if len(tlsSettings) > 0 {
		resp.TLSSettings = tlsSettings
	}

	SuccessWithETag(c, resp)
}

// AgentReport accepts consolidated traffic, alive IPs and status data.
// POST /api/internal/agent/:node_id/report
func AgentReport(c *gin.Context) {
	nodeID := c.Param("node_id")
	var req agentReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	node := resolveNode(nodeID)
	if node == nil {
		NotFound(c, "node not found")
		return
	}

	recorded, err := recordXboardTraffic(node, req.Traffic)
	if err != nil {
		log.Printf("[agent] failed to record report traffic for node %s: %v", nodeID, err)
		InternalError(c, "failed to record traffic")
		return
	}
	aliveProcessed := recordAliveIPs(node, req.Alive)
	markNodeHeartbeatWithStatus(node, req.Status)

	Success(c, gin.H{
		"traffic_recorded": recorded,
		"alive_processed":  aliveProcessed,
		"pull_interval":    database.GetSettingInt("node_pull_interval", 60),
	})
}

// AgentReportTraffic records traffic data from the agent (delta mode).
// POST /api/internal/agent/:node_id/traffic
// Supports both formats:
//   - Xboard-style: {"data": {"1": [upload, download]}}
//   - Legacy: [{"user_uuid":"xxx", "upload":N, "download":N}]
func AgentReportTraffic(c *gin.Context) {
	nodeID := c.Param("node_id")

	// Try Xboard-style format first.
	var xboardReq agentTrafficXboard
	if err := c.ShouldBindBodyWith(&xboardReq, binding.JSON); err == nil && len(xboardReq.Data) > 0 {
		processXboardTraffic(c, nodeID, xboardReq.Data)
		return
	}

	// Try single entry format.
	var req agentTrafficEntry
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err == nil && req.UserUUID != "" {
		processTrafficEntries(c, nodeID, []agentTrafficEntry{req})
		return
	}

	// Try array format for backwards compatibility.
	var entries []agentTrafficEntry
	if err := c.ShouldBindBodyWith(&entries, binding.JSON); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}
	processTrafficEntries(c, nodeID, entries)
}

func recordAgentTraffic(tx *gorm.DB, user model.User, nodeID uint, upload, download int64, recordedAt time.Time) (bool, error) {
	if upload < 0 || download < 0 || (upload == 0 && download == 0) {
		return false, nil
	}

	trafficLog := model.TrafficLog{
		UserID:     user.ID,
		NodeID:     nodeID,
		Upload:     upload,
		Download:   download,
		RecordedAt: recordedAt,
	}
	if err := tx.Create(&trafficLog).Error; err != nil {
		return false, err
	}

	delta := upload + download
	if err := tx.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"upload_used":   gorm.Expr("upload_used + ?", upload),
		"download_used": gorm.Expr("download_used + ?", download),
		"traffic_used":  gorm.Expr("traffic_used + ?", delta),
	}).Error; err != nil {
		return false, err
	}

	if delta > 0 {
		if err := tx.Model(&model.Node{}).Where("id = ?", nodeID).
			UpdateColumn("traffic_used", gorm.Expr("traffic_used + ?", delta)).Error; err != nil {
			return false, err
		}
	}

	return true, nil
}

func markNodeHeartbeat(node *model.Node) {
	now := time.Now()
	database.DB.Model(node).Updates(map[string]interface{}{
		"online":         true,
		"last_heartbeat": &now,
	})
}

func markNodeHeartbeatWithStatus(node *model.Node, status agentHeartbeatReq) {
	now := time.Now()
	database.DB.Model(node).Updates(map[string]interface{}{
		"online":         true,
		"last_heartbeat": &now,
		"cpu":            status.CPU,
		"mem":            status.Mem,
		"uptime":         status.Uptime,
	})
}

func recordXboardTraffic(node *model.Node, data map[string][2]int64) (int, error) {
	if node == nil || len(data) == 0 {
		return 0, nil
	}
	now := time.Now()
	recorded := 0
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for userIDStr, traffic := range data {
			var user model.User
			if err := tx.Where("uuid = ?", userIDStr).First(&user).Error; err != nil {
				var userID uint
				fmt.Sscanf(userIDStr, "%d", &userID)
				if userID > 0 {
					if err := tx.First(&user, userID).Error; err != nil {
						continue
					}
				} else {
					continue
				}
			}
			ok, err := recordAgentTraffic(tx, user, node.ID, traffic[0], traffic[1], now)
			if err != nil {
				return err
			}
			if ok {
				recorded++
			}
		}
		return nil
	})
	return recorded, err
}

// processXboardTraffic handles Xboard-style traffic format: {"user_id": [upload, download]}
func processXboardTraffic(c *gin.Context, nodeID string, data map[string][2]int64) {
	now := time.Now()
	recorded := 0

	node := resolveNode(nodeID)
	if node == nil {
		NotFound(c, "node not found")
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for userIDStr, traffic := range data {
			// Look up user by UUID (Xboard uses string user_id)
			var user model.User
			if err := tx.Where("uuid = ?", userIDStr).First(&user).Error; err != nil {
				// Try numeric ID
				var userID uint
				fmt.Sscanf(userIDStr, "%d", &userID)
				if userID > 0 {
					if err := tx.First(&user, userID).Error; err != nil {
						continue
					}
				} else {
					continue
				}
			}

			ok, err := recordAgentTraffic(tx, user, node.ID, traffic[0], traffic[1], now)
			if err != nil {
				return err
			}
			if ok {
				recorded++
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("[agent] failed to record traffic for node %s: %v", nodeID, err)
		InternalError(c, "failed to record traffic")
		return
	}

	log.Printf("[agent] recorded traffic for %d users (node %s)", recorded, nodeID)
	Success(c, nil)
}

func processTrafficEntries(c *gin.Context, nodeID string, entries []agentTrafficEntry) {
	now := time.Now()
	recorded := 0

	// Resolve node (support custom_id or database id)
	node := resolveNode(nodeID)
	if node == nil {
		NotFound(c, "node not found")
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			// Look up user by UUID
			var user model.User
			if err := tx.Where("uuid = ?", entry.UserUUID).First(&user).Error; err != nil {
				continue // skip unknown users
			}

			ok, err := recordAgentTraffic(tx, user, node.ID, entry.Upload, entry.Download, now)
			if err != nil {
				return err
			}
			if ok {
				recorded++
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("[agent] failed to record traffic for node %s: %v", nodeID, err)
		InternalError(c, "failed to record traffic")
		return
	}

	log.Printf("[agent] recorded traffic for %d users (node %s)", recorded, nodeID)
	Success(c, nil)
}

func updateNodeOnlineCount(nodeID uint, staleThreshold time.Time) error {
	var records []model.AliveIP
	if err := database.DB.Where("node_id = ? AND updated_at >= ?", nodeID, staleThreshold).Find(&records).Error; err != nil {
		return err
	}

	uniqueIPs := make(map[string]struct{})
	for _, record := range records {
		var ips []string
		if err := json.Unmarshal([]byte(record.IPs), &ips); err != nil {
			continue
		}
		for _, ip := range ips {
			if ip == "" {
				continue
			}
			uniqueIPs[ip] = struct{}{}
		}
	}

	return database.DB.Model(&model.Node{}).
		Where("id = ?", nodeID).
		Update("online_count", len(uniqueIPs)).Error
}

func recordAliveIPs(node *model.Node, data map[string][]string) int {
	if node == nil {
		return 0
	}
	if len(data) == 0 {
		database.DB.Where("node_id = ?", node.ID).Delete(&model.AliveIP{})
		database.DB.Model(&model.Node{}).Where("id = ?", node.ID).Update("online_count", 0)
		return 0
	}

	now := time.Now()
	processed := 0
	for userUUID, ips := range data {
		var user model.User
		if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
			continue
		}
		ipsJSON, err := json.Marshal(ips)
		if err != nil {
			continue
		}
		var existing model.AliveIP
		err = database.DB.Where("user_id = ? AND node_id = ?", user.ID, node.ID).First(&existing).Error
		if err == nil {
			database.DB.Model(&existing).Updates(map[string]interface{}{
				"ips":        string(ipsJSON),
				"updated_at": now,
			})
		} else {
			database.DB.Create(&model.AliveIP{
				UserID:    user.ID,
				NodeID:    node.ID,
				IPs:       string(ipsJSON),
				UpdatedAt: now,
			})
		}
		processed++
	}
	staleThreshold := now.Add(-120 * time.Second)
	database.DB.Where("updated_at < ?", staleThreshold).Delete(&model.AliveIP{})
	_ = updateNodeOnlineCount(node.ID, staleThreshold)
	return processed
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

	// Resolve node (support custom_id or database id)
	node := resolveNode(nodeID)
	if node == nil {
		NotFound(c, "node not found")
		return
	}

	if len(req.Data) == 0 {
		database.DB.Where("node_id = ?", node.ID).Delete(&model.AliveIP{})
		database.DB.Model(&model.Node{}).Where("id = ?", node.ID).Update("online_count", 0)
		log.Printf("[agent] processed alive data for 0 users (node %s)", nodeID)
		Success(c, nil)
		return
	}

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
		err = database.DB.Where("user_id = ? AND node_id = ?", user.ID, node.ID).First(&existing).Error
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
				NodeID:    node.ID,
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
	if err := updateNodeOnlineCount(node.ID, staleThreshold); err != nil {
		InternalError(c, "failed to update node online count")
		return
	}

	log.Printf("[agent] processed alive data for %d users (node %s)", processed, nodeID)
	Success(c, nil)
}

// AgentGetAliveList returns count of online IPs per user (only for users with device_limit > 0).
// GET /api/internal/agent/alivelist
func AgentGetAliveList(c *gin.Context) {
	now := time.Now()
	staleThreshold := now.Add(-120 * time.Second)
	var records []struct {
		UserUUID string
		IPs      string
	}

	database.DB.Raw(`
		SELECT u.uuid as user_uuid, a.ips as ips
		FROM alive_ips a
		INNER JOIN users u ON u.id = a.user_id
		WHERE u.status = 1
		  AND u.device_limit > 0
		  AND a.updated_at >= ?
		  AND (u.expired_at IS NULL OR u.expired_at > ?)
		  AND (u.traffic_limit = 0 OR u.traffic_used < u.traffic_limit)
	`, staleThreshold, now).Scan(&records)

	userIPs := make(map[string]map[string]struct{})
	for _, record := range records {
		var ips []string
		if err := json.Unmarshal([]byte(record.IPs), &ips); err != nil {
			continue
		}
		if _, ok := userIPs[record.UserUUID]; !ok {
			userIPs[record.UserUUID] = make(map[string]struct{})
		}
		for _, ip := range ips {
			if ip == "" {
				continue
			}
			userIPs[record.UserUUID][ip] = struct{}{}
		}
	}

	alive := make(map[string]int, len(userIPs))
	for uuid, ips := range userIPs {
		alive[uuid] = len(ips)
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

// resolveNode looks up a node by custom_id first, falling back to database id.
// Returns nil if not found.
func resolveNode(id string) *model.Node {
	var node model.Node
	if err := database.DB.Where("custom_id = ?", id).First(&node).Error; err == nil {
		return &node
	}
	if err := database.DB.First(&node, id).Error; err == nil {
		return &node
	}
	return nil
}

// parseNodeID parses a node ID string to uint, returns 0 on error.
func parseNodeID(id string) uint {
	var n uint
	fmt.Sscanf(id, "%d", &n)
	return n
}

// buildRoutesForNode creates route rules for the node (Xboard-style).
func parseJSONObject(s string) map[string]interface{} {
	if s == "" {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil
	}
	return m
}

// getStringField extracts a string value from a map by trying multiple keys.
func getStringField(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case string:
				return val
			case float64:
				return fmt.Sprintf("%v", val)
			}
		}
	}
	return ""
}

// getIntField extracts an int value from a map by trying multiple keys.
func getIntField(m map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case float64:
				return int(val)
			case int:
				return val
			case string:
				n, err := strconv.Atoi(val)
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}

// extractTLSSettings builds the TLS settings map from network_settings and node fields.
func extractTLSSettings(netSettings map[string]interface{}, security, flowControl string, configJSON ...map[string]interface{}) map[string]interface{} {
	settings := make(map[string]interface{})

	// Helper: get string from netSettings first, then fallback to configJSON
	getWithFallback := func(keys ...string) string {
		if v := getStringField(netSettings, keys...); v != "" {
			return v
		}
		for _, cj := range configJSON {
			if v := getStringField(cj, keys...); v != "" {
				return v
			}
		}
		return ""
	}
	getIntWithFallback := func(keys ...string) int {
		if v := getIntField(netSettings, keys...); v > 0 {
			return v
		}
		for _, cj := range configJSON {
			if v := getIntField(cj, keys...); v > 0 {
				return v
			}
		}
		return 0
	}

	// Reality settings
	if security == "reality" {
		reality := make(map[string]interface{})
		if pk := getWithFallback("reality_private_key", "private_key"); pk != "" {
			reality["private_key"] = pk
		}
		if sid := getWithFallback("reality_short_id", "short_id"); sid != "" {
			reality["short_id"] = sid
		}

		handshake := make(map[string]interface{})
		if hs := getWithFallback("reality_server_name", "server_name", "handshake_server"); hs != "" {
			handshake["server"] = hs
		}
		if hp := getIntWithFallback("reality_port", "handshake_port"); hp > 0 {
			handshake["server_port"] = hp
		}
		if len(handshake) > 0 {
			reality["handshake"] = handshake
		}

		if len(reality) > 0 {
			settings["reality"] = reality
		}
	}

	// TLS server_name
	if sn := getWithFallback("server_name", "tls_server_name"); sn != "" {
		settings["server_name"] = sn
	}

	// Allow insecure
	if ai := getStringField(netSettings, "allow_insecure"); ai == "true" || ai == "1" {
		settings["allow_insecure"] = true
	}

	// ALPN
	if alpn := getStringField(netSettings, "alpn"); alpn != "" {
		settings["alpn"] = alpn
	}

	return settings
}

func buildCustomOutboundsForNode(node model.Node) []customOutboundResp {
	var out []customOutboundResp

	if node.CustomOutbounds != "" {
		var inline []customOutboundResp
		if err := json.Unmarshal([]byte(node.CustomOutbounds), &inline); err == nil {
			out = append(out, inline...)
		}
	}

	var bindings []model.NodeOutbound
	if err := database.DB.Where("node_id = ?", node.ID).Find(&bindings).Error; err != nil || len(bindings) == 0 {
		return out
	}

	ids := make([]uint, 0, len(bindings))
	for _, b := range bindings {
		ids = append(ids, b.CustomOutboundID)
	}

	var customOutbounds []model.CustomOutbound
	database.DB.Where("id IN ? AND status = 1", ids).Order("sort ASC, id ASC").Find(&customOutbounds)
	for _, co := range customOutbounds {
		settings := parseJSONObject(co.SettingsJSON)
		out = append(out, customOutboundResp{
			Tag:      co.Tag,
			Protocol: co.Protocol,
			Settings: settings,
			ProxyTag: co.ProxyTag,
		})
	}
	return out
}

func buildRoutesForNode(node model.Node) []routeRuleResp {
	var rules []model.RouteRule
	database.DB.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&rules)

	result := make([]routeRuleResp, 0, len(rules))
	for _, r := range rules {
		// Parse match string into array
		matchArr := strings.Split(r.Match, "\n")
		if r.Match == "" {
			matchArr = []string{}
		}

		result = append(result, routeRuleResp{
			ID:          int(r.ID),
			Name:        r.Name,
			Match:       matchArr,
			MatchRule:   parseJSONObject(r.MatchJSON),
			Action:      r.Action,
			ActionValue: r.ActionValue,
			ActionRule:  parseJSONObject(r.ActionJSON),
		})
	}
	return result
}
