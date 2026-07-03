package subscription

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"nexus/internal/model"
)

// GenerateV2RayN 生成 Base64 编码的 URI 列表（兼容 V2RayN、ClashR 等客户端）。
func GenerateV2RayN(nodes []model.Node, user model.User) ([]byte, error) {
	var lines []string
	var firstURI string

	for _, node := range nodes {
		if node.Status != 1 {
			continue
		}
		params := ParseNodeParams(node.ConfigJSON)
		var uri string

		switch strings.ToLower(node.Protocol) {
		case "vless":
			uri = buildVlessURI(node, user, params)
		case "hysteria2":
			uri = buildHysteria2URI(node, user, params)
		case "tuic":
			uri = buildTuicURI(node, user, params)
		}
		if uri != "" {
			if firstURI == "" {
				firstURI = uri
			}
			lines = append(lines, uri)
		}
	}

	// 杩藉姞淇℃伅鑺傜偣
	if firstURI != "" {
		expiryName, trafficName := GetInfoNodeNames(user)
		// 淇℃伅鑺傜偣1锛氬埌鏈熸椂闂达級
		info1 := buildInfoURI(firstURI, expiryName)
		lines = append(lines, info1)
		// 淇℃伅鑺傜偣2锛氬墿浣欐祦閲忥級
		info2 := buildInfoURI(firstURI, trafficName)
		lines = append(lines, info2)
	}

	content := strings.Join(lines, "\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	return []byte(encoded), nil
}

// GenerateShadowrocket 生成 Shadowrocket 兼容的 Base64 格式（与 V2RayN 一致）。
func GenerateShadowrocket(nodes []model.Node, user model.User) ([]byte, error) {
	return GenerateV2RayN(nodes, user)
}

// GenerateSurge 生成 Surge 配置格式。
func GenerateSurge(nodes []model.Node, user model.User) ([]byte, error) {
	var b strings.Builder

	b.WriteString("[General]\n")
	b.WriteString("loglevel = notify\n")
	b.WriteString("skip-proxy = 127.0.0.1, 192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12, 100.64.0.0/10, 17.0.0.0/8, localhost, *.local\n")
	b.WriteString("internet-test-url = http://cp.cloudflare.com/generate_204\n")
	b.WriteString("proxy-test-url = http://cp.cloudflare.com/generate_204\n")
	b.WriteString("test-timeout = 3\n")
	b.WriteString("\n[Proxy]\n")

	proxyNames := make([]string, 0)
	var firstLine string

	for _, node := range nodes {
		if node.Status != 1 {
			continue
		}
		params := ParseNodeParams(node.ConfigJSON)
		var line string

		switch strings.ToLower(node.Protocol) {
		case "vless":
			line = buildSurgeVLESS(node, user, params)
		case "hysteria2":
			line = buildSurgeHysteria2(node, user, params)
		case "tuic":
			line = buildSurgeTUIC(node, user, params)
		default:
			continue
		}

		if line != "" {
			if firstLine == "" {
				firstLine = line
			}
			b.WriteString(line + "\n")
			proxyNames = append(proxyNames, node.Name)
		}
	}

	// 杩藉姞淇℃伅鑺傜偣
	if firstLine != "" {
		expiryName, trafficName := GetInfoNodeNames(user)
		info1 := buildSurgeInfoNode(firstLine, expiryName)
		b.WriteString(info1 + "\n")
		proxyNames = append(proxyNames, expiryName)
		info2 := buildSurgeInfoNode(firstLine, trafficName)
		b.WriteString(info2 + "\n")
		proxyNames = append(proxyNames, trafficName)
	}

	// Proxy Group
	b.WriteString("\n[Proxy Group]\n")
	if len(proxyNames) > 0 {
		b.WriteString(fmt.Sprintf("Proxy = select, %s\n", strings.Join(proxyNames, ", ")))
	}

	// Rule
	b.WriteString("\n[Rule]\n")
	b.WriteString("DOMAIN-SUFFIX,google.com,Proxy\n")
	b.WriteString("DOMAIN-SUFFIX,github.com,Proxy\n")
	b.WriteString("DOMAIN-SUFFIX,githubusercontent.com,Proxy\n")
	b.WriteString("DOMAIN-KEYWORD,google,Proxy\n")
	b.WriteString("GEOIP,CN,DIRECT\n")
	b.WriteString("FINAL,Proxy\n")

	return []byte(b.String()), nil
}

// GenerateSurfboard 生成 Surfboard 配置格式。
func GenerateSurfboard(nodes []model.Node, user model.User) ([]byte, error) {
	var b strings.Builder

	b.WriteString("[General]\n")
	b.WriteString("loglevel = notify\n")
	b.WriteString("skip-proxy = 127.0.0.1, 192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12, 100.64.0.0/10, localhost, *.local\n")
	b.WriteString("internet-test-url = http://cp.cloudflare.com/generate_204\n")
	b.WriteString("proxy-test-url = http://cp.cloudflare.com/generate_204\n")
	b.WriteString("test-timeout = 3\n")
	b.WriteString("\n[Proxy]\n")

	proxyNames := make([]string, 0)

	for _, node := range nodes {
		if node.Status != 1 {
			continue
		}
		params := ParseNodeParams(node.ConfigJSON)
		var line string

		switch strings.ToLower(node.Protocol) {
		case "vless":
			line = buildSurgeVLESS(node, user, params) // Surfboard 格式与 Surge 类似
		case "hysteria2":
			line = buildSurgeHysteria2(node, user, params)
		case "tuic":
			line = buildSurgeTUIC(node, user, params)
		default:
			continue
		}

		if line != "" {
			b.WriteString(line + "\n")
			proxyNames = append(proxyNames, node.Name)
		}
	}

	// Proxy Group
	b.WriteString("\n[Proxy Group]\n")
	if len(proxyNames) > 0 {
		b.WriteString(fmt.Sprintf("Proxy = select, %s\n", strings.Join(proxyNames, ", ")))
	}

	// Rule
	b.WriteString("\n[Rule]\n")
	b.WriteString("DOMAIN-SUFFIX,google.com,Proxy\n")
	b.WriteString("DOMAIN-SUFFIX,github.com,Proxy\n")
	b.WriteString("DOMAIN-SUFFIX,githubusercontent.com,Proxy\n")
	b.WriteString("DOMAIN-KEYWORD,google,Proxy\n")
	b.WriteString("GEOIP,CN,DIRECT\n")
	b.WriteString("FINAL,Proxy\n")

	return []byte(b.String()), nil
}

// ==================== URI 构建 ====================

func buildVlessURI(node model.Node, user model.User, p NodeParams) string {
	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}

	q := url.Values{}
	q.Set("flow", "xtls-rprx-vision")
	q.Set("security", "tls")
	q.Set("sni", sni)
	q.Set("type", "tcp")
	q.Set("fp", "chrome")

	if p.PublicKey != "" && p.ShortID != "" {
		q.Set("pbk", p.PublicKey)
		q.Set("sid", p.ShortID)
		if p.HandshakeHost != "" {
			q.Set("sni", p.HandshakeHost)
		}
	}

	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, node.Address, node.Port, q.Encode(), url.QueryEscape(node.Name))
}

func buildHysteria2URI(node model.Node, user model.User, p NodeParams) string {
	password := user.UUID
	if len(password) > 32 {
		password = password[:32]
	}

	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}

	q := url.Values{}
	q.Set("security", "tls")
	q.Set("sni", sni)

	if p.ObfsPassword != "" {
		q.Set("obfs", "salamander")
		q.Set("obfs-password", p.ObfsPassword)
	}

	return fmt.Sprintf("hysteria2://%s@%s:%d?%s#%s",
		password, node.Address, node.Port, q.Encode(), url.QueryEscape(node.Name))
}

func buildTuicURI(node model.Node, user model.User, p NodeParams) string {
	password := user.UUID
	if len(password) > 32 {
		password = password[:32]
	}

	congestion := p.CongestionCtrl
	if congestion == "" {
		congestion = "cubic"
	}

	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}

	q := url.Values{}
	q.Set("congestion_control", congestion)
	q.Set("security", "tls")
	q.Set("sni", sni)

	return fmt.Sprintf("tuic://%s:%s@%s:%d?%s#%s",
		user.UUID, password, node.Address, node.Port, q.Encode(), url.QueryEscape(node.Name))
}

// ==================== Surge 格式构建 ====================

func buildSurgeVLESS(node model.Node, user model.User, p NodeParams) string {
	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}

	line := fmt.Sprintf("%s = vless, %s, %d, uuid=%s, tls=true, sni=%s",
		node.Name, node.Address, node.Port, user.UUID, sni)

	if p.PublicKey != "" {
		line += fmt.Sprintf(", reality-public-key=%s", p.PublicKey)
		if p.ShortID != "" {
			line += fmt.Sprintf(", reality-short-id=%s", p.ShortID)
		}
	}

	return line
}

func buildSurgeHysteria2(node model.Node, user model.User, p NodeParams) string {
	password := user.UUID
	if len(password) > 32 {
		password = password[:32]
	}

	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}

	line := fmt.Sprintf("%s = hysteria2, %s, %d, password=%s, sni=%s",
		node.Name, node.Address, node.Port, password, sni)

	if p.UpMbps > 0 {
		line += fmt.Sprintf(", upload-bandwidth=%d Mbps", p.UpMbps)
	}

	return line
}

func buildSurgeTUIC(node model.Node, user model.User, p NodeParams) string {
	password := user.UUID
	if len(password) > 32 {
		password = password[:32]
	}

	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}

	return fmt.Sprintf("%s = tuic, %s, %d, uuid=%s, password=%s, sni=%s",
		node.Name, node.Address, node.Port, user.UUID, password, sni)
}

// buildInfoURI creates an info node URI (for V2RayN/Shadowrocket format)
func buildInfoURI(firstURI string, infoName string) string {
	idx := strings.LastIndex(firstURI, "#")
	if idx >= 0 {
		return firstURI[:idx] + "#" + url.QueryEscape(infoName)
	}
	return firstURI + "#" + url.QueryEscape(infoName)
}

// buildSurgeInfoNode creates a Surge/Surfboard format info node
func buildSurgeInfoNode(firstLine string, infoName string) string {
	idx := strings.Index(firstLine, " = ")
	if idx >= 0 {
		return infoName + firstLine[idx:]
	}
	return firstLine
}
