package subscription

import (
	"fmt"
	"strings"

	"nexus/internal/model"

	"gopkg.in/yaml.v3"
)

// clashConfig 是 Clash/Clash.Meta 配置文件的顶层结构。
type clashConfig struct {
	MixedPort    int              `yaml:"mixed-port"`
	AllowLan     bool             `yaml:"allow-lan"`
	Mode         string           `yaml:"mode"`
	LogLevel     string           `yaml:"log-level"`
	Proxies      []interface{}    `yaml:"proxies"`
	ProxyGroups  []clashGroup     `yaml:"proxy-groups"`
	Rules        []string         `yaml:"rules"`
}

type clashGroup struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Proxies []string `yaml:"proxies"`
}

// GenerateClash 生成 Clash/Clash.Meta 格式的 YAML 配置。
func GenerateClash(nodes []model.Node, user model.User) ([]byte, error) {
	cfg := clashConfig{
		MixedPort: 7890,
		AllowLan:  false,
		Mode:      "rule",
		LogLevel:  "info",
		Proxies:   make([]interface{}, 0),
		Rules: []string{
			"DOMAIN-SUFFIX,google.com,Proxy",
			"DOMAIN-SUFFIX,github.com,Proxy",
			"DOMAIN-SUFFIX,githubusercontent.com,Proxy",
			"DOMAIN-SUFFIX,googleapis.com,Proxy",
			"DOMAIN-SUFFIX,ggpht.com,Proxy",
			"DOMAIN-KEYWORD,google,Proxy",
			"GEOIP,LAN,DIRECT",
			"GEOIP,CN,DIRECT",
			"MATCH,Proxy",
		},
	}

	nodeNames := make([]string, 0)

	for _, node := range nodes {
		if node.Status != 1 {
			continue
		}
		params := ParseNodeParams(node.ConfigJSON)
		var proxy map[string]interface{}

		switch strings.ToLower(node.Protocol) {
		case "vless":
			proxy = buildClashVLESS(node, user, params)
		case "hysteria2":
			proxy = buildClashHysteria2(node, user, params)
		case "tuic":
			proxy = buildClashTUIC(node, user, params)
		default:
			continue
		}

		if proxy != nil {
			cfg.Proxies = append(cfg.Proxies, proxy)
			nodeNames = append(nodeNames, node.Name)
		}
	}

	// 构建代理分组
	if len(nodeNames) > 0 {
		allProxies := make([]string, len(nodeNames))
		copy(allProxies, nodeNames)

		cfg.ProxyGroups = []clashGroup{
			{
				Name:    "Proxy",
				Type:    "select",
				Proxies: allProxies,
			},
		}
	}

	return yaml.Marshal(cfg)
}

func buildClashVLESS(node model.Node, user model.User, p NodeParams) map[string]interface{} {
	proxy := map[string]interface{}{
		"name":          node.Name,
		"type":          "vless",
		"server":        node.Address,
		"port":          node.Port,
		"uuid":          user.UUID,
		"flow":          "xtls-rprx-vision",
		"network":       "tcp",
		"tls":           true,
		"udp":           true,
	}

	if p.ServerName != "" {
		proxy["servername"] = p.ServerName
	}

	// Reality 配置
	if p.PublicKey != "" && p.ShortID != "" {
		realityOpts := map[string]interface{}{
			"public-key": p.PublicKey,
			"short-id":   p.ShortID,
		}
		if p.HandshakeHost != "" {
			realityOpts["handshake"] = p.HandshakeHost
			if p.HandshakePort > 0 {
				realityOpts["handshake-port"] = p.HandshakePort
			}
		}
		proxy["reality-opts"] = realityOpts
		proxy["client-fingerprint"] = "chrome"
	}

	return proxy
}

func buildClashHysteria2(node model.Node, user model.User, p NodeParams) map[string]interface{} {
	password := user.UUID
	if len(password) > 32 {
		password = password[:32]
	}

	proxy := map[string]interface{}{
		"name":     node.Name,
		"type":     "hysteria2",
		"server":   node.Address,
		"port":     node.Port,
		"password": password,
		"tls":      true,
		"udp":      true,
	}

	if p.ServerName != "" {
		proxy["sni"] = p.ServerName
	}

	if p.UpMbps > 0 {
		proxy["up"] = fmt.Sprintf("%d Mbps", p.UpMbps)
	}
	if p.DownMbps > 0 {
		proxy["down"] = fmt.Sprintf("%d Mbps", p.DownMbps)
	}

	if p.ObfsPassword != "" {
		proxy["obfs"] = map[string]interface{}{
			"type":     "salamander",
			"password": p.ObfsPassword,
		}
	}

	return proxy
}

func buildClashTUIC(node model.Node, user model.User, p NodeParams) map[string]interface{} {
	password := user.UUID
	if len(password) > 32 {
		password = password[:32]
	}

	congestion := p.CongestionCtrl
	if congestion == "" {
		congestion = "cubic"
	}

	proxy := map[string]interface{}{
		"name":              node.Name,
		"type":              "tuic",
		"server":            node.Address,
		"port":              node.Port,
		"uuid":              user.UUID,
		"password":          password,
		"congestion-control": congestion,
		"tls":               true,
		"udp":               true,
	}

	if p.ServerName != "" {
		proxy["sni"] = p.ServerName
	}

	return proxy
}
