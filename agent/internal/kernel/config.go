package kernel

import (
	"encoding/json"
	"fmt"
	"strings"
)

// NodeConfig represents the node configuration received from the panel.
type NodeConfig struct {
	ConfigMode        string                 `json:"config_mode,omitempty"` // "auto" or "json"
	ConfigJSON        string                 `json:"config_json,omitempty"` // raw config for json mode
	NodeID            int                    `json:"node_id,omitempty"`
	Protocol          string                 `json:"protocol"`
	ListenIP          string                 `json:"listen_ip"`
	ServerPort        int                    `json:"server_port"`
	StatsPort         int                    `json:"stats_port,omitempty"`
	Network           string                 `json:"network"`
	NetworkSettings   map[string]interface{} `json:"networkSettings,omitempty"`
	BaseConfig        BaseConfig             `json:"base_config"`
	Routes            []RouteRule            `json:"routes,omitempty"`
	KernelType        string                 `json:"kernel_type,omitempty"`
	CertConfig        CertConfig             `json:"cert_config,omitempty"`
	CustomOutbounds   []CustomOutbound       `json:"custom_outbounds,omitempty"`
	CertPEM           string                 `json:"-"`
	KeyPEM            string                 `json:"-"`

	// TLS
	TLS         int                    `json:"tls,omitempty"`
	Flow        string                 `json:"flow,omitempty"`
	TLSSettings map[string]interface{} `json:"tls_settings,omitempty"`

	// VLESS Reality
	ServerName string `json:"server_name,omitempty"`

	// Hysteria
	Version      int    `json:"version,omitempty"`
	UpMbps       int    `json:"up_mbps,omitempty"`
	DownMbps     int    `json:"down_mbps,omitempty"`
	Obfs         string `json:"obfs,omitempty"`
	ObfsPassword string `json:"obfs-password,omitempty"`

	// TUIC
	CongestionControl string `json:"congestion_control,omitempty"`
}

type BaseConfig struct {
	PushInterval int `json:"push_interval"`
	PullInterval int `json:"pull_interval"`
}

type RouteRule struct {
	ID          int                    `json:"id"`
	Match       []string               `json:"match"`
	MatchRule   map[string]interface{} `json:"match_rule,omitempty"`
	Action      string                 `json:"action"`
	ActionValue string                 `json:"action_value,omitempty"`
	ActionRule  map[string]interface{} `json:"action_rule,omitempty"`
}

type CertConfig struct {
	CertMode    string            `json:"cert_mode"`
	Domain      string            `json:"domain"`
	Email       string            `json:"email"`
	DNSProvider string            `json:"dns_provider"`
	DNSEnv      map[string]string `json:"dns_env"`
	HTTPPort    int               `json:"http_port"`
	CertFile    string            `json:"cert_file"`
	KeyFile     string            `json:"key_file"`
	CertContent string            `json:"cert_content"`
	KeyContent  string            `json:"key_content"`
	CertDir     string            `json:"cert_dir"`
}

type CustomOutbound struct {
	Tag      string                 `json:"tag"`
	Protocol string                 `json:"protocol"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	ProxyTag string                 `json:"proxy_tag,omitempty"`
}

// User represents a user for sing-box configuration.
type User struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	SpeedLimit  int    `json:"speed_limit"`
	DeviceLimit int    `json:"device_limit"`
}

// NodeConfigFromPanel converts an httpclient.NodeConfigResponse to kernel.NodeConfig.
type NodeConfigFromPanel struct {
	ConfigMode        string
	ConfigJSON        string
	NodeID            int
	Protocol          string
	ListenIP          string
	ServerPort        int
	Network           string
	NetworkSettings   map[string]interface{}
	BaseConfig        BaseConfig
	Routes            []RouteRule
	KernelType        string
	CertConfig        CertConfig
	CustomOutbounds   []CustomOutbound
	CertPEM           string
	KeyPEM            string
	TLS               int
	Flow              string
	TLSSettings       map[string]interface{}
	ServerName        string
	UpMbps            int
	DownMbps          int
	ObfsPassword      string
	CongestionControl string
}

// ToNodeConfig converts the panel response to kernel NodeConfig.
func (p *NodeConfigFromPanel) ToNodeConfig() NodeConfig {
	return NodeConfig{
		ConfigMode:        p.ConfigMode,
		ConfigJSON:        p.ConfigJSON,
		NodeID:            p.NodeID,
		Protocol:          p.Protocol,
		ListenIP:          p.ListenIP,
		ServerPort:        p.ServerPort,
		Network:           p.Network,
		NetworkSettings:   p.NetworkSettings,
		BaseConfig:        p.BaseConfig,
		Routes:            p.Routes,
		KernelType:        p.KernelType,
		CertConfig:        p.CertConfig,
		CustomOutbounds:   p.CustomOutbounds,
		CertPEM:           p.CertPEM,
		KeyPEM:            p.KeyPEM,
		TLS:               p.TLS,
		Flow:              p.Flow,
		TLSSettings:       p.TLSSettings,
		ServerName:        p.ServerName,
		UpMbps:            p.UpMbps,
		DownMbps:          p.DownMbps,
		ObfsPassword:      p.ObfsPassword,
		CongestionControl: p.CongestionControl,
	}
}

// SingboxConfig is the complete sing-box configuration.
type SingboxConfig struct {
	Log          logConfig                `json:"log"`
	DNS          dnsConfig                `json:"dns"`
	Inbounds     []any                    `json:"inbounds"`
	Outbounds    []map[string]interface{} `json:"outbounds"`
	Route        routeConfig              `json:"route"`
	Experimental experimentalConfig       `json:"experimental"`
}

type dnsConfig struct {
	Servers         []dnsServer      `json:"servers"`
	Rules           []dnsRule        `json:"rules"`
	Final           string           `json:"final"`
	Strategy        string           `json:"strategy,omitempty"`
	FakeIP          *fakeIPConfig    `json:"fakeip,omitempty"`
	IndependentCache bool            `json:"independent_cache"`
}

type dnsServer struct {
	Tag     string `json:"tag"`
	Address string `json:"address"`
	Detour  string `json:"detour,omitempty"`
}

type dnsRule struct {
	Outbound []string `json:"outbound,omitempty"`
	Server   string   `json:"server"`
	Inbound  []string `json:"inbound,omitempty"`
	Rule     string   `json:"rule,omitempty"`
}

type fakeIPConfig struct {
	Enabled    bool   `json:"enabled"`
	Inet4Range string `json:"inet4_range"`
	Inet6Range string `json:"inet6_range"`
}

type logConfig struct {
	Level     string `json:"level"`
	Timestamp bool   `json:"timestamp"`
}

type outbound struct {
	Type string `json:"type"`
	Tag  string `json:"tag"`
}

type routeConfig struct {
	Rules []map[string]interface{} `json:"rules"`
	Final string                   `json:"final"`
}

type experimentalConfig struct {
	CacheFile cacheFileConfig `json:"cache_file"`
	ClashAPI  clashAPIConfig  `json:"clash_api"`
}

type cacheFileConfig struct {
	Enabled bool `json:"enabled"`
}

type clashAPIConfig struct {
	ExternalController string `json:"external_controller"`
}

// GenerateSingboxConfig generates a complete sing-box configuration from node parameters and users.
func GenerateSingboxConfig(nodeConfig NodeConfig, users []User) (string, error) {
	cfg := baseConfig(nodeConfig)

	switch strings.ToLower(nodeConfig.Protocol) {
	case "vless":
		inbound, err := buildVLESSInbound(nodeConfig, users)
		if err != nil {
			return "", err
		}
		cfg.Inbounds = append(cfg.Inbounds, inbound)

	case "hysteria2", "hy2":
		inbound, err := buildHysteria2Inbound(nodeConfig, users)
		if err != nil {
			return "", err
		}
		cfg.Inbounds = append(cfg.Inbounds, inbound)

	case "tuic":
		inbound, err := buildTUICInbound(nodeConfig, users)
		if err != nil {
			return "", err
		}
		cfg.Inbounds = append(cfg.Inbounds, inbound)

	default:
		return "", fmt.Errorf("unsupported protocol: %s", nodeConfig.Protocol)
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func baseConfig(nodeConfig NodeConfig) SingboxConfig {
		outbounds := []map[string]interface{}{
			{"type": "direct", "tag": "direct"},
			{"type": "block", "tag": "block"},
			{"type": "dns", "tag": "dns-out"},
		}
		outbounds = append(outbounds, buildCustomOutbounds(nodeConfig.CustomOutbounds)...)

		rules := []map[string]interface{}{
			{"inbound": []string{"dns-in"}, "outbound": "dns-out"},
		}
		rules = append(rules, buildRouteRules(nodeConfig.Routes)...)

		statsPort := nodeConfig.StatsPort
		if statsPort == 0 {
			statsPort = 9090
		}

		return SingboxConfig{
			Log: logConfig{
				Level:     "info",
				Timestamp: true,
			},
			DNS: buildDNSConfig(),
			Outbounds: outbounds,
			Route: routeConfig{
				Rules: rules,
				Final: "direct",
			},
			Experimental: experimentalConfig{
				CacheFile: cacheFileConfig{Enabled: true},
				ClashAPI: clashAPIConfig{
					ExternalController: fmt.Sprintf("127.0.0.1:%d", statsPort),
				},
			},
		}
	}

	func buildDNSConfig() dnsConfig {
		return dnsConfig{
			Servers: []dnsServer{
				{
					Tag:     "dns-remote",
					Address: "https://1.1.1.1/dns-query",
					Detour:  "direct",
				},
				{
					Tag:     "dns-google",
					Address: "https://dns.google/dns-query",
					Detour:  "direct",
				},
				{
					Tag:     "dns-fakeip",
					Address: "fakeip",
				},
			},
			Rules: []dnsRule{
				{
					Outbound: []string{"any"},
					Server:   "dns-remote",
				},
			},
			Final:           "dns-remote",
			IndependentCache: true,
			FakeIP: &fakeIPConfig{
				Enabled:    true,
				Inet4Range: "198.18.0.0/15",
				Inet6Range: "fc00::/18",
			},
		}
	}

func buildCustomOutbounds(items []CustomOutbound) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if item.Tag == "" || item.Protocol == "" {
			continue
		}
		m := map[string]interface{}{
			"type": item.Protocol,
			"tag":  item.Tag,
		}
		for k, v := range item.Settings {
			m[k] = v
		}
		if item.ProxyTag != "" {
			m["detour"] = item.ProxyTag
		}
		out = append(out, m)
	}
	return out
}

func buildRouteRules(items []RouteRule) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		rule := map[string]interface{}{}
		for k, v := range item.MatchRule {
			rule[k] = v
		}
		if len(rule) == 0 {
			rule = legacyMatchToRule(item.Match)
		}
		if len(rule) == 0 {
			continue
		}
		outbound := routeOutbound(item)
		if outbound == "" {
			outbound = "direct"
		}
		rule["outbound"] = outbound
		out = append(out, rule)
	}
	return out
}

func legacyMatchToRule(lines []string) map[string]interface{} {
	rule := map[string]interface{}{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key := "domain_suffix"
		value := line
		if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
			key = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])
		}
		if value == "" {
			continue
		}
		field := routeMatchField(key)
		appendRuleValue(rule, field, value)
	}
	return rule
}

func routeMatchField(key string) string {
	switch strings.ToLower(key) {
	case "domain", "full":
		return "domain"
	case "suffix", "domain_suffix":
		return "domain_suffix"
	case "keyword", "domain_keyword":
		return "domain_keyword"
	case "ip", "cidr", "ip_cidr":
		return "ip_cidr"
	case "source_ip", "source_cidr":
		return "source_ip_cidr"
	case "port":
		return "port"
	case "protocol":
		return "protocol"
	default:
		return "domain_suffix"
	}
}

func appendRuleValue(rule map[string]interface{}, key, value string) {
	if existing, ok := rule[key].([]string); ok {
		rule[key] = append(existing, value)
		return
	}
	rule[key] = []string{value}
}

func routeOutbound(item RouteRule) string {
	if v, ok := item.ActionRule["target"].(string); ok && v != "" {
		return v
	}
	if v, ok := item.ActionRule["outbound"].(string); ok && v != "" {
		return v
	}
	switch strings.ToLower(item.Action) {
	case "block", "reject":
		return "block"
	case "direct":
		return "direct"
	case "route", "proxy":
		return item.ActionValue
	default:
		if item.ActionValue != "" {
			return item.ActionValue
		}
		return item.Action
	}
}

// VLESS + Reality structures
type vlessInbound struct {
	Type       string      `json:"type"`
	Tag        string      `json:"tag"`
	Listen     string      `json:"listen"`
	ListenPort int         `json:"listen_port"`
	Users      []vlessUser `json:"users"`
	TLS        vlessTLS    `json:"tls"`
}

type vlessUser struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

type vlessTLS struct {
	Enabled    bool         `json:"enabled"`
	ServerName string       `json:"server_name"`
	Reality    vlessReality `json:"reality"`
}

type vlessReality struct {
	Enabled    bool             `json:"enabled"`
	Handshake  realityHandshake `json:"handshake"`
	PrivateKey string           `json:"private_key"`
	ShortID    []string         `json:"short_id"`
}

type realityHandshake struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
}

func buildVLESSInbound(nodeConfig NodeConfig, users []User) (vlessInbound, error) {
	tlsSettings := nodeConfig.TLSSettings
	if tlsSettings == nil {
		tlsSettings = make(map[string]interface{})
	}

	// Extract Reality settings
	realitySettings, _ := tlsSettings["reality"].(map[string]interface{})
	privateKey := ""
	shortID := ""
	handshakeServer := nodeConfig.ServerName
	handshakePort := 443

	if realitySettings != nil {
		if pk, ok := realitySettings["private_key"].(string); ok {
			privateKey = pk
		}
		if sid, ok := realitySettings["short_id"].(string); ok {
			shortID = sid
		}
		if hs, ok := realitySettings["handshake"].(map[string]interface{}); ok {
			if s, ok := hs["server"].(string); ok {
				handshakeServer = s
			}
			if p, ok := hs["server_port"].(float64); ok {
				handshakePort = int(p)
			}
		}
	}

	// Build users
	vUsers := make([]vlessUser, len(users))
	for i, u := range users {
		name := u.UUID
		if len(name) > 8 {
			name = name[:8]
		}
		vUsers[i] = vlessUser{Name: name, UUID: u.UUID}
	}

	serverName := nodeConfig.ServerName
	if serverName == "" {
		if sn, ok := tlsSettings["server_name"].(string); ok {
			serverName = sn
		}
	}

	return vlessInbound{
		Type:       "vless",
		Tag:        "vless-reality",
		Listen:     nodeConfig.ListenIP,
		ListenPort: nodeConfig.ServerPort,
		Users:      vUsers,
		TLS: vlessTLS{
			Enabled:    true,
			ServerName: serverName,
			Reality: vlessReality{
				Enabled: true,
				Handshake: realityHandshake{
					Server:     handshakeServer,
					ServerPort: handshakePort,
				},
				PrivateKey: privateKey,
				ShortID:    []string{shortID},
			},
		},
	}, nil
}

// Hysteria2 structures
type hysteria2Inbound struct {
	Type       string          `json:"type"`
	Tag        string          `json:"tag"`
	Listen     string          `json:"listen"`
	ListenPort int             `json:"listen_port"`
	Users      []hysteria2User `json:"users"`
	TLS        hysteria2TLS    `json:"tls"`
	Bandwidth  *bandwidth      `json:"bandwidth,omitempty"`
	Obfs       *obfsConfig     `json:"obfs,omitempty"`
}

type hysteria2User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	UpMbps   int    `json:"up_mbps,omitempty"`
	DownMbps int    `json:"down_mbps,omitempty"`
}

type hysteria2TLS struct {
	Enabled         bool     `json:"enabled"`
	ServerName      string   `json:"server_name,omitempty"`
	CertificatePath string   `json:"certificate_path,omitempty"`
	KeyPath         string   `json:"key_path,omitempty"`
	Certificate     []string `json:"certificate,omitempty"`
	Key             []string `json:"key,omitempty"`
}

type bandwidth struct {
	Up   string `json:"up"`
	Down string `json:"down"`
}

type obfsConfig struct {
	Type     string `json:"type"`
	Password string `json:"password"`
}

func buildHysteria2Inbound(nodeConfig NodeConfig, users []User) (hysteria2Inbound, error) {
	// Build users
	hyUsers := make([]hysteria2User, len(users))
	for i, u := range users {
		pw := strings.ReplaceAll(u.UUID, "-", "")
		if len(pw) > 32 {
			pw = pw[:32]
		}
		hy := hysteria2User{Name: pw[:8], Password: pw}
		if u.SpeedLimit > 0 {
			hy.UpMbps = u.SpeedLimit
			hy.DownMbps = u.SpeedLimit
		}
		hyUsers[i] = hy
	}

	in := hysteria2Inbound{
		Type:       "hysteria2",
		Tag:        "hysteria2-in",
		Listen:     nodeConfig.ListenIP,
		ListenPort: nodeConfig.ServerPort,
		Users:      hyUsers,
		TLS: hysteria2TLS{
			Enabled: true,
		},
	}
	applyCertificateToHysteria2TLS(&in.TLS, nodeConfig)

	if nodeConfig.UpMbps > 0 || nodeConfig.DownMbps > 0 {
		in.Bandwidth = &bandwidth{
			Up:   fmt.Sprintf("%d mbps", nodeConfig.UpMbps),
			Down: fmt.Sprintf("%d mbps", nodeConfig.DownMbps),
		}
	}

	if nodeConfig.ObfsPassword != "" {
		in.Obfs = &obfsConfig{
			Type:     "salamander",
			Password: nodeConfig.ObfsPassword,
		}
	}

	return in, nil
}

// TUIC structures
type tuicInbound struct {
	Type              string     `json:"type"`
	Tag               string     `json:"tag"`
	Listen            string     `json:"listen"`
	ListenPort        int        `json:"listen_port"`
	Users             []tuicUser `json:"users"`
	CongestionControl string     `json:"congestion_control"`
	TLS               tuicTLS    `json:"tls"`
}

type tuicUser struct {
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
	Password string `json:"password"`
}

type tuicTLS struct {
	Enabled         bool     `json:"enabled"`
	ServerName      string   `json:"server_name,omitempty"`
	CertificatePath string   `json:"certificate_path,omitempty"`
	KeyPath         string   `json:"key_path,omitempty"`
	Certificate     []string `json:"certificate,omitempty"`
	Key             []string `json:"key,omitempty"`
}

func buildTUICInbound(nodeConfig NodeConfig, users []User) (tuicInbound, error) {
	// Build users
	tUsers := make([]tuicUser, len(users))
	for i, u := range users {
		name := u.UUID
		if len(name) > 8 {
			name = name[:8]
		}
		tUsers[i] = tuicUser{
			Name:     name,
			UUID:     u.UUID,
			Password: u.UUID,
		}
	}

	congestion := nodeConfig.CongestionControl
	if congestion == "" {
		congestion = "cubic"
	}

	in := tuicInbound{
		Type:              "tuic",
		Tag:               "tuic-in",
		Listen:            nodeConfig.ListenIP,
		ListenPort:        nodeConfig.ServerPort,
		Users:             tUsers,
		CongestionControl: congestion,
		TLS: tuicTLS{
			Enabled: true,
		},
	}
	applyCertificateToTUICTLS(&in.TLS, nodeConfig)
	return in, nil
}

func applyCertificateToHysteria2TLS(tls *hysteria2TLS, nodeConfig NodeConfig) {
	if nodeConfig.ServerName != "" {
		tls.ServerName = nodeConfig.ServerName
	} else if nodeConfig.CertConfig.Domain != "" {
		tls.ServerName = nodeConfig.CertConfig.Domain
	}
	if nodeConfig.CertPEM != "" && nodeConfig.KeyPEM != "" {
		tls.Certificate = []string{nodeConfig.CertPEM}
		tls.Key = []string{nodeConfig.KeyPEM}
	}
}

func applyCertificateToTUICTLS(tls *tuicTLS, nodeConfig NodeConfig) {
	if nodeConfig.ServerName != "" {
		tls.ServerName = nodeConfig.ServerName
	} else if nodeConfig.CertConfig.Domain != "" {
		tls.ServerName = nodeConfig.CertConfig.Domain
	}
	if nodeConfig.CertPEM != "" && nodeConfig.KeyPEM != "" {
		tls.Certificate = []string{nodeConfig.CertPEM}
		tls.Key = []string{nodeConfig.KeyPEM}
	}
}
