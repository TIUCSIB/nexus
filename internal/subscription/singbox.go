package subscription

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"nexus/internal/model"
)

// NodeParams 解析节点 ConfigJSON / NetworkSettings 中的连接参数。
type NodeParams struct {
	ServerName     string `json:"server_name"`
	PrivateKey     string `json:"private_key"`
	PublicKey      string `json:"public_key"`
	ShortID        string `json:"short_id"`
	HandshakeHost  string `json:"handshake_host"`
	HandshakePort  int    `json:"handshake_port"`
	UpMbps         int    `json:"up_mbps"`
	DownMbps       int    `json:"down_mbps"`
	ObfsPassword   string `json:"obfs_password"`
	ObfsType       string `json:"obfs_type"`
	CongestionCtrl string `json:"congestion_control"`
	AllowInsecure  bool   `json:"allow_insecure"`
	ObfsEnabled    bool   `json:"obfs_enabled"`
}

func parseJSONMap(raw string) map[string]interface{} {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil
	}
	return data
}

func getNestedMapField(data map[string]interface{}, key string) map[string]interface{} {
	if data == nil {
		return nil
	}

	value, ok := data[key]
	if !ok || value == nil {
		return nil
	}

	result, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	return result
}

func getStringField(data map[string]interface{}, keys ...string) string {
	if data == nil {
		return ""
	}

	for _, key := range keys {
		value, ok := data[key]
		if !ok || value == nil {
			continue
		}

		switch v := value.(type) {
		case string:
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				return trimmed
			}
		}
	}

	return ""
}

func getIntField(data map[string]interface{}, keys ...string) int {
	if data == nil {
		return 0
	}

	for _, key := range keys {
		value, ok := data[key]
		if !ok || value == nil {
			continue
		}

		switch v := value.(type) {
		case int:
			return v
		case int8:
			return int(v)
		case int16:
			return int(v)
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float32:
			return int(v)
		case float64:
			return int(v)
		case json.Number:
			if i, err := v.Int64(); err == nil {
				return int(i)
			}
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				return parsed
			}
		}
	}

	return 0
}

func getBoolField(data map[string]interface{}, keys ...string) (bool, bool) {
	if data == nil {
		return false, false
	}

	for _, key := range keys {
		value, ok := data[key]
		if !ok || value == nil {
			continue
		}

		switch v := value.(type) {
		case bool:
			return v, true
		case float64:
			return v != 0, true
		case float32:
			return v != 0, true
		case int:
			return v != 0, true
		case int8:
			return v != 0, true
		case int16:
			return v != 0, true
		case int32:
			return v != 0, true
		case int64:
			return v != 0, true
		case string:
			trimmed := strings.TrimSpace(strings.ToLower(v))
			switch trimmed {
			case "1", "true", "yes", "on":
				return true, true
			case "0", "false", "no", "off":
				return false, true
			}
		}
	}

	return false, false
}

// ParseNodeParams 从节点的 ConfigJSON 和 NetworkSettings 字段解析连接参数。
func ParseNodeParams(configJSON string, networkSettings string) NodeParams {
	var p NodeParams

	configJSONMap := parseJSONMap(configJSON)
	networkSettingsMap := parseJSONMap(networkSettings)

	// Step 1: 先读取 config_json 的扁平字段。
	p.ServerName = getStringField(configJSONMap, "server_name", "tls_server_name", "reality_server_name", "handshake_server")
	p.PrivateKey = getStringField(configJSONMap, "private_key")
	p.PublicKey = getStringField(configJSONMap, "public_key")
	p.ShortID = getStringField(configJSONMap, "short_id")
	p.HandshakeHost = getStringField(configJSONMap, "handshake_host", "handshake_server")
	p.HandshakePort = getIntField(configJSONMap, "handshake_port")
	p.UpMbps = getIntField(configJSONMap, "up_mbps", "bandwidth_up")
	p.DownMbps = getIntField(configJSONMap, "down_mbps", "bandwidth_down")
	p.ObfsPassword = getStringField(configJSONMap, "obfs_password", "obfs-password")
	p.ObfsType = getStringField(configJSONMap, "obfs_type")
	p.CongestionCtrl = getStringField(configJSONMap, "congestion_control")
	if allowInsecure, ok := getBoolField(configJSONMap, "allow_insecure", "tls_allow_insecure"); ok {
		p.AllowInsecure = allowInsecure
	}
	if obfsEnabled, ok := getBoolField(configJSONMap, "obfs_open"); ok {
		p.ObfsEnabled = obfsEnabled
	}

	// Step 1b: 兼容 config_json 中的嵌套结构。
	if tlsSettings := getNestedMapField(configJSONMap, "tls_settings"); tlsSettings != nil {
		if p.ServerName == "" {
			p.ServerName = getStringField(tlsSettings, "server_name", "tls_server_name")
		}
		if allowInsecure, ok := getBoolField(tlsSettings, "allow_insecure", "tls_allow_insecure"); ok {
			p.AllowInsecure = allowInsecure
		}
		if reality := getNestedMapField(tlsSettings, "reality"); reality != nil {
			if p.PublicKey == "" {
				p.PublicKey = getStringField(reality, "public_key")
			}
			if p.PrivateKey == "" {
				p.PrivateKey = getStringField(reality, "private_key")
			}
			if p.ShortID == "" {
				p.ShortID = getStringField(reality, "short_id")
			}
			if p.HandshakeHost == "" {
				p.HandshakeHost = getStringField(reality, "server_name", "handshake_host", "dest", "server")
			}
			if p.HandshakePort == 0 {
				p.HandshakePort = getIntField(reality, "server_port", "handshake_port", "port")
			}
		}
	}
	if bandwidth := getNestedMapField(configJSONMap, "bandwidth"); bandwidth != nil {
		if p.UpMbps == 0 {
			p.UpMbps = getIntField(bandwidth, "up")
		}
		if p.DownMbps == 0 {
			p.DownMbps = getIntField(bandwidth, "down")
		}
	}
	if obfs := getNestedMapField(configJSONMap, "obfs"); obfs != nil {
		if obfsEnabled, ok := getBoolField(obfs, "open"); ok {
			p.ObfsEnabled = obfsEnabled
		}
		if p.ObfsType == "" {
			p.ObfsType = getStringField(obfs, "type")
		}
		if p.ObfsPassword == "" {
			p.ObfsPassword = getStringField(obfs, "password", "obfs_password", "obfs-password")
		}
	}

	// Step 2: network_settings 优先级更高，按 Agent 侧兼容逻辑覆盖。
	if serverName := getStringField(networkSettingsMap, "server_name", "tls_server_name", "reality_server_name"); serverName != "" {
		p.ServerName = serverName
	}
	if privateKey := getStringField(networkSettingsMap, "reality_private_key", "private_key"); privateKey != "" {
		p.PrivateKey = privateKey
	}
	if publicKey := getStringField(networkSettingsMap, "reality_public_key", "public_key"); publicKey != "" {
		p.PublicKey = publicKey
	}
	if shortID := getStringField(networkSettingsMap, "reality_short_id", "short_id"); shortID != "" {
		p.ShortID = shortID
	}
	if handshakeHost := getStringField(networkSettingsMap, "reality_server_name", "handshake_host", "handshake_server"); handshakeHost != "" {
		p.HandshakeHost = handshakeHost
	}
	if handshakePort := getIntField(networkSettingsMap, "reality_port", "handshake_port"); handshakePort > 0 {
		p.HandshakePort = handshakePort
	}
	if upMbps := getIntField(networkSettingsMap, "bandwidth_up", "up_mbps"); upMbps > 0 {
		p.UpMbps = upMbps
	}
	if downMbps := getIntField(networkSettingsMap, "bandwidth_down", "down_mbps"); downMbps > 0 {
		p.DownMbps = downMbps
	}
	if obfsPassword := getStringField(networkSettingsMap, "obfs_password", "obfs-password"); obfsPassword != "" {
		p.ObfsPassword = obfsPassword
	}
	if obfsType := getStringField(networkSettingsMap, "obfs_type"); obfsType != "" {
		p.ObfsType = obfsType
	}
	if congestionControl := getStringField(networkSettingsMap, "congestion_control"); congestionControl != "" {
		p.CongestionCtrl = congestionControl
	}
	if allowInsecure, ok := getBoolField(networkSettingsMap, "allow_insecure", "tls_allow_insecure"); ok {
		p.AllowInsecure = allowInsecure
	}
	if obfsEnabled, ok := getBoolField(networkSettingsMap, "obfs_open"); ok {
		p.ObfsEnabled = obfsEnabled
	}

	if p.ServerName == "" && p.HandshakeHost != "" {
		p.ServerName = p.HandshakeHost
	}
	if p.ObfsPassword != "" {
		p.ObfsEnabled = true
		if p.ObfsType == "" {
			p.ObfsType = "salamander"
		}
	}

	return p
}

// GenerateSingbox 生成 sing-box 客户端配置 JSON。
// 遍历所有节点，根据协议类型生成对应的 outbound，
// 同时附带路由配置（direct + block）。
func GenerateSingbox(nodes []model.Node, user model.User) ([]byte, error) {
	generatedOutbounds := make([]json.RawMessage, 0, len(nodes))

	for _, node := range nodes {
		if node.Status != 1 {
			continue
		}
		params := ParseNodeParams(node.ConfigJSON, node.NetworkSettings)
		var ob json.RawMessage
		var err error

		switch strings.ToLower(node.Protocol) {
		case "vless":
			ob, err = buildSingboxVLESS(node, user, params)
		case "hysteria2":
			ob, err = buildSingboxHysteria2(node, user, params)
		case "tuic":
			ob, err = buildSingboxTUIC(node, user, params)
		default:
			continue
		}
		if err != nil {
			continue
		}
		generatedOutbounds = append(generatedOutbounds, ob)
	}

	templateContent := GetSubscriptionTemplate(SettingSubscribeTemplateSingbox)
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(templateContent), &config); err != nil {
		return nil, err
	}

	outbounds, ok := config["outbounds"].([]interface{})
	if !ok {
		return nil, errors.New("Sing-box 模板的 outbounds 格式无效")
	}

	proxyTags := collectSingboxTags(generatedOutbounds)
	config["outbounds"] = injectSingboxOutbounds(outbounds, generatedOutbounds, proxyTags)

	return json.MarshalIndent(config, "", "  ")
}

func collectSingboxTags(outbounds []json.RawMessage) []string {
	tags := make([]string, 0, len(outbounds))
	for _, ob := range outbounds {
		var m map[string]interface{}
		if err := json.Unmarshal(ob, &m); err != nil {
			continue
		}
		if tag, ok := m["tag"].(string); ok && tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func injectSingboxOutbounds(templateOutbounds []interface{}, generatedOutbounds []json.RawMessage, proxyTags []string) []interface{} {
	result := make([]interface{}, 0, len(templateOutbounds)+len(generatedOutbounds))

	for _, generated := range generatedOutbounds {
		var outbound map[string]interface{}
		if err := json.Unmarshal(generated, &outbound); err == nil {
			result = append(result, outbound)
		}
	}

	for _, item := range templateOutbounds {
		obj, ok := item.(map[string]interface{})
		if ok {
			if outboundsValue, exists := obj["outbounds"]; exists {
				if outboundNames, ok := outboundsValue.([]interface{}); ok {
					obj["outbounds"] = replaceSingboxOutboundPlaceholders(outboundNames, proxyTags)
				}
			}
		}
		result = append(result, item)
	}

	return result
}

func replaceSingboxOutboundPlaceholders(outbounds []interface{}, proxyTags []string) []interface{} {
	result := make([]interface{}, 0, len(outbounds)+len(proxyTags))
	for _, item := range outbounds {
		name, ok := item.(string)
		if !ok {
			result = append(result, item)
			continue
		}
		if name == singboxAutoOutboundsPlaceholder {
			for _, tag := range proxyTags {
				result = append(result, tag)
			}
			continue
		}
		result = append(result, item)
	}
	return result
}

func buildSingboxVLESS(node model.Node, user model.User, p NodeParams) (json.RawMessage, error) {
	ob := map[string]interface{}{
		"type":        "vless",
		"tag":         node.Name,
		"server":      node.Address,
		"server_port": node.Port,
		"uuid":        user.UUID,
	}

	// Flow
	flow := node.FlowControl
	if flow != "" && flow != "none" {
		ob["flow"] = flow
	}

	// UDP 包编码：plain TCP 无 TLS 时不强制 xudp，避免客户端误显示
	security := strings.ToLower(strings.TrimSpace(node.Security))
	if security == "" {
		security = "none"
	}
	hasTLS := security == "tls" || security == "reality"
	if hasTLS {
		// Reality/TLS + VLESS 常用 xudp
		ob["packet_encoding"] = "xudp"
	}

	// TLS 配置 — 仅当 security 为 tls 或 reality 时启用
	if !hasTLS {
		return json.Marshal(ob)
	}

	serverName := p.ServerName
	if serverName == "" {
		serverName = p.HandshakeHost
	}
	tlsConfig := map[string]interface{}{
		"enabled": true,
	}
	if serverName != "" {
		tlsConfig["server_name"] = serverName
	}
	if p.AllowInsecure {
		tlsConfig["insecure"] = true
	}
	if security == "reality" && p.PublicKey != "" {
		reality := map[string]interface{}{
			"enabled":    true,
			"public_key": p.PublicKey,
		}
		if p.ShortID != "" {
			reality["short_id"] = p.ShortID
		}
		if p.HandshakeHost != "" {
			hp := p.HandshakeHost
			port := p.HandshakePort
			if port == 0 {
				port = 443
			}
			reality["handshake"] = map[string]interface{}{
				"server":      hp,
				"server_port": port,
			}
		}
		tlsConfig["reality"] = reality
		tlsConfig["utls"] = map[string]interface{}{
			"enabled":     true,
			"fingerprint": "chrome",
		}
	}
	ob["tls"] = tlsConfig

	return json.Marshal(ob)
}

func buildSingboxHysteria2(node model.Node, user model.User, p NodeParams) (json.RawMessage, error) {
	// Hysteria2 使用 UUID 前 32 字符作为密码
	password := user.UUID
	if len(password) > 32 {
		password = password[:32]
	}

	ob := map[string]interface{}{
		"type":        "hysteria2",
		"tag":         node.Name,
		"server":      node.Address,
		"server_port": node.Port,
		"password":    password,
	}

	// TLS 配置
	tlsConfig := map[string]interface{}{
		"enabled": true,
	}
	if p.ServerName != "" {
		tlsConfig["server_name"] = p.ServerName
	}
	if p.AllowInsecure {
		tlsConfig["insecure"] = true
	}
	ob["tls"] = tlsConfig

	// 带宽配置
	if p.UpMbps > 0 {
		ob["up_mbps"] = p.UpMbps
	}
	if p.DownMbps > 0 {
		ob["down_mbps"] = p.DownMbps
	}

	// 混淆
	if p.ObfsEnabled && p.ObfsPassword != "" {
		obfsType := p.ObfsType
		if obfsType == "" {
			obfsType = "salamander"
		}
		ob["obfs"] = map[string]interface{}{
			"type":     obfsType,
			"password": p.ObfsPassword,
		}
	}

	return json.Marshal(ob)
}

func buildSingboxTUIC(node model.Node, user model.User, p NodeParams) (json.RawMessage, error) {
	password := user.UUID
	if len(password) > 32 {
		password = password[:32]
	}

	congestion := p.CongestionCtrl
	if congestion == "" {
		congestion = "cubic"
	}

	ob := map[string]interface{}{
		"type":               "tuic",
		"tag":                node.Name,
		"server":             node.Address,
		"server_port":        node.Port,
		"uuid":               user.UUID,
		"password":           password,
		"congestion_control": congestion,
	}

	// TLS 配置
	tlsConfig := map[string]interface{}{
		"enabled": true,
	}
	if p.ServerName != "" {
		tlsConfig["server_name"] = p.ServerName
	}
	if p.AllowInsecure {
		tlsConfig["insecure"] = true
	}
	ob["tls"] = tlsConfig

	return json.Marshal(ob)
}
