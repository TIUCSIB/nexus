package subscription

import (
	"encoding/json"
	"strings"

	"nexus/internal/model"
)

// NodeParams 解析节点 ConfigJSON 中的连接参数。
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
	CongestionCtrl string `json:"congestion_control"`
}

// ParseNodeParams 从节点的 ConfigJSON 和 NetworkSettings 字段解析连接参数。
func ParseNodeParams(configJSON string, networkSettings string) NodeParams {
	var p NodeParams
	if configJSON == "" {
		return p
	}
	_ = json.Unmarshal([]byte(configJSON), &p)
	// NetworkSettings 中的字段覆盖 ConfigJSON 的默认值
	if networkSettings != "" {
		_ = json.Unmarshal([]byte(networkSettings), &p)
	}
	return p
}

// GenerateSingbox 生成 sing-box 客户端配置 JSON。
// 遍历所有节点，根据协议类型生成对应的 outbound，
// 同时附带路由配置（direct + block）。
func GenerateSingbox(nodes []model.Node, user model.User) ([]byte, error) {
	outbounds := make([]json.RawMessage, 0, len(nodes)+3)

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
		outbounds = append(outbounds, ob)
	}

	// 如果没有可用节点，仍然返回一个有效配置（只有 direct 和 block）
	autoSelectTag := ""
	proxyTags := make([]string, 0)
	for i, ob := range outbounds {
		var m map[string]interface{}
		_ = json.Unmarshal(ob, &m)
		if tag, ok := m["tag"].(string); ok {
			proxyTags = append(proxyTags, tag)
			if i == 0 {
				autoSelectTag = tag
			}
		}
	}

	// 添加 urltest（自动选择）outbound
	if len(proxyTags) > 0 {
		urltest := map[string]interface{}{
			"type":    "urltest",
			"tag":     "proxy",
			"outbounds": proxyTags,
			"url":     "https://www.gstatic.com/generate_204",
			"interval": "3m",
		}
		urltestBytes, _ := json.Marshal(urltest)
		outbounds = append(outbounds, urltestBytes)
		_ = autoSelectTag
	}

	// direct outbound
	direct := map[string]interface{}{
		"type": "direct",
		"tag":  "direct",
	}
	directBytes, _ := json.Marshal(direct)
	outbounds = append(outbounds, directBytes)

	// block outbound
	block := map[string]interface{}{
		"type": "block",
		"tag":  "block",
	}
	blockBytes, _ := json.Marshal(block)
	outbounds = append(outbounds, blockBytes)

// dns outbound
		dnsOB := map[string]interface{}{
			"type": "dns",
			"tag":  "dns-out",
		}
		dnsBytes, _ := json.Marshal(dnsOB)
		outbounds = append(outbounds, dnsBytes)

		// 构建完整配置
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level": "info",
		},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{"tag": "google", "address": "tls://8.8.8.8"},
				{"tag": "local", "address": "local"},
			},
			"rules": []map[string]interface{}{
				{"outbound": []string{"any"}, "server": "local"},
			},
		},
		"inbounds": []map[string]interface{}{
			{
				"type":    "mixed",
				"tag":     "mixed-in",
				"listen":  "127.0.0.1",
				"listen_port": 2080,
			},
		},
		"outbounds": outbounds,
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"protocol": "dns",
					"outbound": "dns-out",
				},
				{
					"ip_is_private": true,
					"outbound":      "direct",
				},
			},
			"final":        "proxy",
			"auto_detect_interface": true,
		},
	}

	return json.MarshalIndent(config, "", "  ")
}

func buildSingboxVLESS(node model.Node, user model.User, p NodeParams) (json.RawMessage, error) {
	flow := node.FlowControl
	if flow == "" {
		flow = "none"
	}
	ob := map[string]interface{}{
		"type":        "vless",
		"tag":         node.Name,
		"server":      node.Address,
		"server_port": node.Port,
		"uuid":        user.UUID,
		"flow":        flow,
	}

	// TLS 配置
	tlsConfig := map[string]interface{}{
		"enabled": true,
		"server_name": p.ServerName,
	}
	if p.PublicKey != "" && p.ShortID != "" {
		reality := map[string]interface{}{
			"enabled":    true,
			"public_key": p.PublicKey,
			"short_id":   p.ShortID,
		}
		if p.HandshakeHost != "" {
			hp := p.HandshakeHost
			if p.HandshakePort > 0 {
				reality["handshake"] = map[string]interface{}{
					"server": hp,
					"server_port": p.HandshakePort,
				}
			} else {
				reality["handshake"] = map[string]interface{}{
					"server": hp,
					"server_port": 443,
				}
			}
		}
		tlsConfig["reality"] = reality
	}
	ob["tls"] = tlsConfig

	// transport
	ob["transport"] = map[string]interface{}{
		"type": "tcp",
	}

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
		"enabled":     true,
		"server_name": p.ServerName,
	}
	ob["tls"] = tlsConfig

	// 带宽配置
	if p.UpMbps > 0 && p.DownMbps > 0 {
		ob["up_mbps"] = p.UpMbps
		ob["down_mbps"] = p.DownMbps
	}

	// 混淆
	if p.ObfsPassword != "" {
		ob["obfs"] = map[string]interface{}{
			"type":     "salamander",
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
		"type":        "tuic",
		"tag":         node.Name,
		"server":      node.Address,
		"server_port": node.Port,
		"uuid":        user.UUID,
		"password":    password,
		"congestion_control": congestion,
	}

	// TLS 配置
	tlsConfig := map[string]interface{}{
		"enabled":     true,
		"server_name": p.ServerName,
	}
	ob["tls"] = tlsConfig

	return json.Marshal(ob)
}
