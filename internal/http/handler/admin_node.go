package handler

import (
	"encoding/json"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/curve25519"
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
	Sort            int     `json:"sort"`
	Status          *int    `json:"status"`
	NetworkSettings string  `json:"network_settings"`
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
		CustomID:  req.CustomID,
		Name:            req.Name,
		Address:         req.Address,
		Protocol:        req.Protocol,
		Port:            req.Port,
			ServicePort:     req.ServicePort,
		GroupID:         req.GroupID,
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
		NetworkSettings: req.NetworkSettings,
		RegisterToken:   uuid.New().String(),
		Sort:            req.Sort,
		Status:          status,
	}

	if err := database.DB.Create(&node).Error; err != nil {
		InternalError(c, "创建节点失败")
		return
	}

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
	Sort            *int     `json:"sort"`
	Status          *int     `json:"status"`
	NetworkSettings string   `json:"network_settings"`
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
		updates["group_id"] = *req.GroupID
	}
	if req.RouteID != nil {
		updates["route_id"] = *req.RouteID
	}
	if req.Rate != nil {
		updates["rate"] = *req.Rate
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.TrafficLimit != nil {
		updates["traffic_limit"] = *req.TrafficLimit
	}
	if req.TrafficUsed != nil {
		updates["traffic_used"] = *req.TrafficUsed
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
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
	if req.NetworkSettings != "" {
		updates["network_settings"] = req.NetworkSettings
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
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
	Success(c, gin.H{"message": "流量已重置"})
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
			"up_mbps":      100,
			"down_mbps":    500,
			"obfs_password": "",
			"cert_path":    "/etc/nexus/cert.pem",
			"key_path":     "/etc/nexus/key.pem",
		}
	case "tuic":
		params = map[string]interface{}{
			"congestion_control": "cubic",
			"cert_path":         "/etc/nexus/cert.pem",
			"key_path":          "/etc/nexus/key.pem",
		}
	case "vmess":
		params = map[string]interface{}{
			"security":  "auto",
			"encryption": "none",
			"alter_id":  0,
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
