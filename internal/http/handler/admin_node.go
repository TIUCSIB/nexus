package handler

import (
	"encoding/json"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/ws"

	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/curve25519"
	"gorm.io/gorm"
)

func AdminListNodes(c *gin.Context) {
	page, pageSize := parsePagination(c)
	groupID := c.Query("group_id")

	var total int64
	query := database.DB.Model(&model.Node{})
	if groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}
	query.Count(&total)

	var nodes []model.Node
	offset := (page - 1) * pageSize
	query.Order("sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&nodes)

	SuccessPage(c, nodes, total, page, pageSize)
}

type createNodeRequest struct {
	CustomID        string  `json:"custom_id"`
	Name            string  `json:"name" binding:"required"`
	Address         string  `json:"address" binding:"required"`
	Protocol        string  `json:"protocol" binding:"required"`
	Port            int     `json:"port" binding:"required"`
	ServicePort     int     `json:"service_port"`
	GroupID         *uint   `json:"group_id"`
	GroupIDs        []uint  `json:"group_ids"`
	RouteID         *uint   `json:"route_id"`
	Rate            float64 `json:"rate"`
	Tags            string  `json:"tags"`
	TrafficLimit    int64   `json:"traffic_limit"`
	ParentID        *uint   `json:"parent_id"`
	Security        string  `json:"security"`
	Transport       string  `json:"transport"`
	FlowControl     string  `json:"flow_control"`
	VlessEncryption bool    `json:"vless_encryption"`
	ConfigMode      string  `json:"config_mode"`
	ConfigJSON      string  `json:"config_json"`
	CertConfig      string  `json:"cert_config"`
	KernelType      string  `json:"kernel_type"`
CustomOutbounds string  `json:"custom_outbounds"`
		Sort            int     `json:"sort"`
		Status          *int    `json:"status"`
		NetworkSettings string  `json:"network_settings"`
		MachineID       *uint   `json:"machine_id"`
	}

func AdminCreateNode(c *gin.Context) {
	var req createNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请填写节点名称、地址、协议和端口")
		return
	}

	configMode := req.ConfigMode
	if configMode == "" {
		configMode = "auto"
	}

	rate := req.Rate
	if rate <= 0 {
		rate = 1
	}

	security := req.Security
	if security == "" {
		security = "none"
	}

	transport := req.Transport
	if transport == "" {
		transport = "tcp"
	}

	flowControl := req.FlowControl
	if flowControl == "" {
		flowControl = "none"
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	// 如果是自动模式但没有 ConfigJSON，尝试从 Name/Protocol 信息构建一个默认的
	configJSON := req.ConfigJSON
	if configMode == "auto" && configJSON == "" {
		defaultParams := buildDefaultConfigJSON(req.Protocol)
		configJSON = defaultParams
	}

	node := model.Node{
		CustomID:        req.CustomID,
		Name:            req.Name,
		Address:         req.Address,
		Protocol:        req.Protocol,
		Port:            req.Port,
		ServicePort:     req.ServicePort,
		GroupID:         req.GroupID,
		GroupIDs:        req.GroupIDs,
		RouteID:         req.RouteID,
		Rate:            rate,
		Tags:            req.Tags,
		TrafficLimit:    req.TrafficLimit,
		ParentID:        req.ParentID,
		Security:        security,
		Transport:       req.Transport,
		FlowControl:     flowControl,
		VlessEncryption: req.VlessEncryption,
		ConfigMode:      configMode,
		ConfigJSON:      configJSON,
		CertConfig:      req.CertConfig,
		KernelType:      defaultKernelType(req.KernelType),
CustomOutbounds: req.CustomOutbounds,
	NetworkSettings: req.NetworkSettings,
	MachineID:       req.MachineID,
	RegisterToken:   uuid.New().String(),
		Sort:            req.Sort,
		Status:          status,
	}

	if err := database.DB.Create(&node).Error; err != nil {
		InternalError(c, "创建节点失败")
		return
	}

	recordAudit(c, "create_node", fmt.Sprintf("node:%d", node.ID), detailJSON(gin.H{"name": node.Name}))
	Success(c, gin.H{
		"node":           node,
		"register_token": node.RegisterToken,
	})
}

type updateNodeRequest struct {
	CustomID        *string  `json:"custom_id"`
	Name            string   `json:"name"`
	Address         string   `json:"address"`
	Protocol        string   `json:"protocol"`
	Port            *int     `json:"port"`
	ServicePort     *int     `json:"service_port"`
	GroupID         *uint    `json:"group_id"`
	GroupIDs        []uint   `json:"group_ids"`
	RouteID         *uint    `json:"route_id"`
	Rate            *float64 `json:"rate"`
	Tags            string   `json:"tags"`
	TrafficLimit    *int64   `json:"traffic_limit"`
	TrafficUsed     *int64   `json:"traffic_used"`
	ParentID        *uint    `json:"parent_id"`
	Security        string   `json:"security"`
	Transport       string   `json:"transport"`
	FlowControl     string   `json:"flow_control"`
	VlessEncryption *bool    `json:"vless_encryption"`
	ConfigMode      string   `json:"config_mode"`
	ConfigJSON      string   `json:"config_json"`
	CertConfig      *string  `json:"cert_config"`
	KernelType      string   `json:"kernel_type"`
	CustomOutbounds *string  `json:"custom_outbounds"`
		Sort            *int     `json:"sort"`
		Status          *int     `json:"status"`
		NetworkSettings string   `json:"network_settings"`
		MachineID       *uint    `json:"machine_id"`
	}

func AdminUpdateNode(c *gin.Context) {
	id := c.Param("id")

	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	var req updateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请求参数格式错误")
		return
	}

	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Protocol != "" {
		updates["protocol"] = req.Protocol
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.ServicePort != nil {
		updates["service_port"] = *req.ServicePort
	}
	if req.GroupID != nil {
		if *req.GroupID == 0 {
			updates["group_id"] = nil
		} else {
			updates["group_id"] = *req.GroupID
		}
	}
	if req.GroupIDs != nil {
		// GORM's Updates(map) skips serializer:json, so we must serialize manually
		if groupIDsJSON, err := json.Marshal(req.GroupIDs); err == nil {
			updates["group_ids"] = string(groupIDsJSON)
		}
	}
	if req.RouteID != nil {
		if *req.RouteID == 0 {
			updates["route_id"] = nil
		} else {
			updates["route_id"] = *req.RouteID
		}
	}
	if req.Rate != nil {
		updates["rate"] = *req.Rate
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.CustomID != nil {
		updates["custom_id"] = *req.CustomID
	}
	if req.TrafficLimit != nil {
		updates["traffic_limit"] = *req.TrafficLimit
	}
	if req.TrafficUsed != nil {
		updates["traffic_used"] = *req.TrafficUsed
	}
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			updates["parent_id"] = nil
		} else {
			updates["parent_id"] = *req.ParentID
		}
	}
	if req.Security != "" {
		updates["security"] = req.Security
	}
	if req.Transport != "" {
		updates["transport"] = req.Transport
	}
	if req.FlowControl != "" {
		updates["flow_control"] = req.FlowControl
	}
	if req.VlessEncryption != nil {
		updates["vless_encryption"] = *req.VlessEncryption
	}
	if req.ConfigMode != "" {
		updates["config_mode"] = req.ConfigMode
	}
	if req.ConfigJSON != "" {
		updates["config_json"] = req.ConfigJSON
	}
	if req.CertConfig != nil {
		updates["cert_config"] = *req.CertConfig
	}
	if req.KernelType != "" {
		updates["kernel_type"] = req.KernelType
	}
	if req.CustomOutbounds != nil {
		updates["custom_outbounds"] = *req.CustomOutbounds
	}
	if req.NetworkSettings != "" {
		updates["network_settings"] = req.NetworkSettings
	}
if req.Sort != nil {
			updates["sort"] = *req.Sort
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}
		if req.MachineID != nil {
			if *req.MachineID == 0 {
				updates["machine_id"] = nil
			} else {
				updates["machine_id"] = *req.MachineID
			}
		}

	if len(updates) == 0 {
		BadRequest(c, "没有需要更新的字段")
		return
	}

	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&node).Updates(updates).Error; err != nil {
		InternalError(c, "更新节点失败")
		return
	}

database.DB.First(&node, id)

	// Notify connected agent to reload config
	WSHub.SendCommand(fmt.Sprintf("node:%d", node.ID), &ws.Command{Type: "reload"})

	recordAudit(c, "update_node", fmt.Sprintf("node:%d", node.ID), detailJSON(gin.H{"updated_fields": keys(updates)}))
	Success(c, node)
}

func AdminDeleteNode(c *gin.Context) {
	id := c.Param("id")

	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	if err := database.DB.Delete(&node).Error; err != nil {
		InternalError(c, "删除节点失败")
		return
	}

	recordAudit(c, "delete_node", fmt.Sprintf("node:%d", node.ID), detailJSON(gin.H{"name": node.Name}))
	Success(c, gin.H{"message": "节点已删除"})
}

func AdminRestartNode(c *gin.Context) {
	id := c.Param("id")

	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	if node.Status != 1 {
		BadRequest(c, "节点未启用，无法重启")
		return
	}

	// Touch updated_at so the agent detects a config change on next heartbeat
	// and pulls fresh config, effectively restarting sing-box.
	now := time.Now()
	if err := database.DB.Model(&node).Update("updated_at", &now).Error; err != nil {
		InternalError(c, "发送重启指令失败")
		return
	}

	recordAudit(c, "restart_node", fmt.Sprintf("node:%d", node.ID), detailJSON(gin.H{"name": node.Name}))
	Success(c, gin.H{"message": "重启指令已发送", "node_id": node.ID})
}

func AdminResetNodeTraffic(c *gin.Context) {
	id := c.Param("id")

	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	database.DB.Model(&node).Update("traffic_used", 0)
	recordAudit(c, "reset_node_traffic", fmt.Sprintf("node:%d", node.ID), detailJSON(gin.H{"name": node.Name}))
	Success(c, gin.H{"message": "流量已重置"})
}

// defaultKernelType normalises the kernel_type field, defaulting to "singbox".
func defaultKernelType(s string) string {
	if s == "" {
		return "singbox"
	}
	return s
}

// buildDefaultConfigJSON 为指定协议生成默认的配置参数 JSON
func buildDefaultConfigJSON(protocol string) string {
	var params map[string]interface{}

	switch protocol {
	case "vless":
		params = map[string]interface{}{
			"server_name":      "",
			"private_key":      "",
			"short_id":         "6ba85179e30d4fc2",
			"handshake_server": "www.microsoft.com",
			"handshake_port":   443,
		}
	case "hysteria2", "hy2":
		params = map[string]interface{}{
			"up_mbps":       100,
			"down_mbps":     500,
			"obfs_password": "",
			"cert_path":     "/etc/nexus/cert.pem",
			"key_path":      "/etc/nexus/key.pem",
		}
	case "tuic":
		params = map[string]interface{}{
			"congestion_control": "cubic",
			"cert_path":          "/etc/nexus/cert.pem",
			"key_path":           "/etc/nexus/key.pem",
		}
	case "vmess":
		params = map[string]interface{}{
			"security":   "auto",
			"encryption": "none",
			"alter_id":   0,
		}
	case "trojan":
		params = map[string]interface{}{
			"password": "",
		}
	case "shadowsocks":
		params = map[string]interface{}{
			"method":   "aes-128-gcm",
			"password": "",
		}
	default:
		params = map[string]interface{}{}
	}

	b, err := json.Marshal(params)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func AdminGenerateRealityKeys(c *gin.Context) {
	var privateKey [32]byte
	if _, err := crypto_rand.Read(privateKey[:]); err != nil {
		InternalError(c, "生成密钥失败")
		return
	}
	publicKey, err := curve25519.X25519(privateKey[:], curve25519.Basepoint)
	if err != nil {
		InternalError(c, fmt.Sprintf("生成公钥失败: %v", err))
		return
	}
	privateKeyB64 := base64.RawURLEncoding.EncodeToString(privateKey[:])
	publicKeyB64 := base64.RawURLEncoding.EncodeToString(publicKey)
	Success(c, gin.H{
		"private_key": privateKeyB64,
		"public_key":  publicKeyB64,
	})
}

// ---------- Batch operations (Xboard-style) ----------

// AdminBatchDeleteNodes batch-deletes nodes by IDs.
// POST /api/admin/nodes/batch-delete
func AdminBatchDeleteNodes(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		BadRequest(c, "请提供要删除的节点ID列表")
		return
	}

	if err := database.DB.Where("id IN ?", req.IDs).Delete(&model.Node{}).Error; err != nil {
		InternalError(c, "批量删除失败")
		return
	}

	Success(c, gin.H{"message": "批量删除成功"})
}

// AdminBatchResetNodeTraffic batch-resets traffic for nodes by IDs.
// POST /api/admin/nodes/batch-reset-traffic
func AdminBatchResetNodeTraffic(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		BadRequest(c, "请提供要重置流量的节点ID列表")
		return
	}

	if err := database.DB.Model(&model.Node{}).Where("id IN ?", req.IDs).Update("traffic_used", 0).Error; err != nil {
		InternalError(c, "批量重置流量失败")
		return
	}

	Success(c, gin.H{"message": "批量重置流量成功"})
}

// AdminBatchUpdateNodes batch-updates node properties (show, enabled, machine_id).
// POST /api/admin/nodes/batch-update
func AdminBatchUpdateNodes(c *gin.Context) {
	var req struct {
		IDs       []uint `json:"ids" binding:"required"`
		Show      *int   `json:"show"`
		Enabled   *bool  `json:"enabled"`
		MachineID *uint  `json:"machine_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		BadRequest(c, "请提供要更新的节点ID列表")
		return
	}

	updates := map[string]interface{}{}
	if req.Show != nil {
		updates["show"] = *req.Show
	}
	if req.Enabled != nil {
		if *req.Enabled {
			updates["status"] = 1
		} else {
			updates["status"] = 0
		}
	}
	if req.MachineID != nil {
		if *req.MachineID == 0 {
			updates["machine_id"] = nil
		} else {
			updates["machine_id"] = *req.MachineID
		}
	}

	if len(updates) == 0 {
		BadRequest(c, "没有需要更新的字段，请指定 show/enabled/machine_id 之一")
		return
	}

	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&model.Node{}).Where("id IN ?", req.IDs).Updates(updates).Error; err != nil {
		InternalError(c, "批量更新失败")
		return
	}

	Success(c, gin.H{"message": "批量更新成功"})
}

// AdminSortNodes saves node sort order.
// POST /api/admin/nodes/sort
func AdminSortNodes(c *gin.Context) {
	var req []struct {
		ID    uint `json:"id"`
		Order int  `json:"order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req) == 0 {
		BadRequest(c, "请提供节点排序数据")
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range req {
			if err := tx.Model(&model.Node{}).Where("id = ?", item.ID).Update("sort", item.Order).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		InternalError(c, "保存排序失败")
		return
	}

	Success(c, gin.H{"message": "排序已保存"})
}

// AdminCopyNode duplicates a node (traffic zeroed, show disabled, name suffixed).
// POST /api/admin/nodes/copy
func AdminCopyNode(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请提供要复制的节点ID")
		return
	}

	var original model.Node
	if err := database.DB.First(&original, req.ID).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	copied := original
	copied.ID = 0
	copied.Name = original.Name + " - 副本"
	copied.CustomID = ""
	copied.TrafficUsed = 0
	copied.Status = 0 // disabled by default
	copied.RegisterToken = uuid.New().String()
	copied.CreatedAt = time.Now()
	copied.UpdatedAt = time.Now()

	if err := database.DB.Create(&copied).Error; err != nil {
		InternalError(c, "复制节点失败")
		return
	}

	Success(c, gin.H{"node": copied, "message": "节点已复制"})
}

// AdminGenerateECHKey generates ECH (Encrypted Client Hello) key pair.
// POST /api/admin/nodes/generate-ech-key
func AdminGenerateECHKey(c *gin.Context) {
	publicName := c.DefaultQuery("public_name", "ech.example.com")
	if len(publicName) < 1 || len(publicName) > 253 {
		BadRequest(c, "public_name 长度必须在 1-253 字符之间")
		return
	}

	// Generate X25519 key pair
	privateKey := make([]byte, 32)
	if _, err := crypto_rand.Read(privateKey); err != nil {
		InternalError(c, "生成密钥失败")
		return
	}
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		InternalError(c, "生成公钥失败")
		return
	}

	configID := make([]byte, 1)
	crypto_rand.Read(configID)

	// Build ECHConfigContents (draft-ietf-tls-esni-18)
	contents := []byte{}
	contents = append(contents, configID[0])             // config_id
	contents = append(contents, 0x00, 0x20)              // kem_id: DHKEM(X25519)
	contents = append(contents, 0x00, 0x20)              // public_key length (32)
	contents = append(contents, publicKey...)             // public_key
	// cipher_suites: 2 suites x 4 bytes = 8 bytes
	contents = append(contents, 0x00, 0x08)              // cipher_suites byte length
	contents = append(contents, 0x00, 0x01, 0x00, 0x01) // HKDF-SHA256 + AES-128-GCM
	contents = append(contents, 0x00, 0x01, 0x00, 0x03) // HKDF-SHA256 + ChaCha20Poly1305
	contents = append(contents, 0x00)                    // max_name_length
	contents = append(contents, byte(len(publicName)))   // public_name length
	contents = append(contents, []byte(publicName)...)   // public_name
	contents = append(contents, 0x00, 0x00)              // extensions: empty

	// ECHConfig = version(2) + length(2) + contents
	echConfig := []byte{}
	echConfig = append(echConfig, 0xfe, 0x0d)                        // version
	echConfig = append(echConfig, byte(len(contents)>>8), byte(len(contents))) // length
	echConfig = append(echConfig, contents...)

	// ECHConfigList = total_length(2) + configs
	echConfigList := []byte{}
	echConfigList = append(echConfigList, byte(len(echConfig)>>8), byte(len(echConfig)))
	echConfigList = append(echConfigList, echConfig...)

	// ECH Keys = private_key_len(2) + key(32) + config_len(2) + config
	echKeysPayload := []byte{}
	echKeysPayload = append(echKeysPayload, 0x00, 0x20) // private_key length (32)
	echKeysPayload = append(echKeysPayload, privateKey...)
	echKeysPayload = append(echKeysPayload, byte(len(echConfig)>>8), byte(len(echConfig)))
	echKeysPayload = append(echKeysPayload, echConfig...)

	keyPem := "-----BEGIN ECH KEYS-----\n" + chunkSplit(base64.StdEncoding.EncodeToString(echKeysPayload), 64) + "-----END ECH KEYS-----"
	configPem := "-----BEGIN ECH CONFIGS-----\n" + chunkSplit(base64.StdEncoding.EncodeToString(echConfigList), 64) + "-----END ECH CONFIGS-----"

	Success(c, gin.H{
		"key":    keyPem,
		"config": configPem,
	})
}

// chunkSplit splits a string into chunks of the given length separated by newlines.
func chunkSplit(s string, chunkLen int) string {
	result := ""
	for i := 0; i < len(s); i += chunkLen {
		end := i + chunkLen
		if end > len(s) {
			end = len(s)
		}
		result += s[i:end] + "\n"
	}
	return result
}
