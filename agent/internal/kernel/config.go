package kernel

import (
	"encoding/base64"
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
	Servers          []map[string]interface{} `json:"servers"`
	Rules            []map[string]interface{} `json:"rules,omitempty"`
	Final            string                   `json:"final,omitempty"`
	Strategy         string                   `json:"strategy,omitempty"`
	IndependentCache bool                     `json:"independent_cache,omitempty"`
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
	Rules                []map[string]interface{} `json:"rules"`
	Final                string                   `json:"final"`
	DefaultDomainResolver string                  `json:"default_domain_resolver,omitempty"`
}

type experimentalConfig struct {
	CacheFile cacheFileConfig `json:"cache_file"`
	ClashAPI  clashAPIConfig  `json:"clash_api"`
}

type cacheFileConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path,omitempty"`
}

type clashAPIConfig struct {
	ExternalController string `json:"external_controller"`
}

// listenAddr normalizes the listen IP for sing-box inbound.
// If the configured IP is not a special address (0.0.0.0, 127.0.0.1, ::),
// it defaults to 0.0.0.0 to avoid "cannot assign requested address" errors
// when the public IP isn't directly bound to a network interface.
func listenAddr(ip string) string {
	if ip == "" || ip == "0.0.0.0" || ip == "127.0.0.1" || ip == "::" || ip == "::1" {
		if ip == "" {
			return "0.0.0.0"
		}
		return ip
	}
	// Public/private IPs that might not be bound to the interface — use 0.0.0.0
	return "0.0.0.0"
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
	}
	outbounds = append(outbounds, buildCustomOutbounds(nodeConfig.CustomOutbounds)...)

	// Only route rules from panel; do not reference non-existent dns-in inbound.
	rules := buildRouteRules(nodeConfig.Routes)

	statsPort := nodeConfig.StatsPort
	if statsPort == 0 {
		statsPort = 9090
	}

	return SingboxConfig{
		Log: logConfig{
			Level:     "info",
			Timestamp: true,
		},
		DNS:       buildDNSConfig(),
		Outbounds: outbounds,
		Route: routeConfig{
			Rules:                 rules,
			Final:                 "direct",
			DefaultDomainResolver: "local",
		},
		Experimental: experimentalConfig{
			CacheFile: cacheFileConfig{
				Enabled: false,
			},
			ClashAPI: clashAPIConfig{
				ExternalController: fmt.Sprintf("127.0.0.1:%d", statsPort),
			},
		},
	}
}

func buildDNSConfig() dnsConfig {
	// Minimal DNS for server-side inbound nodes (sing-box 1.12+ compatible).
	return dnsConfig{
		Servers: []map[string]interface{}{
			{
				"type": "local",
				"tag":  "local",
			},
			{
				"type":   "udp",
				"tag":    "google",
				"server": "8.8.8.8",
			},
		},
		Final: "local",
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

// VLESS structures
type vlessInbound struct {
	Type       string       `json:"type"`
	Tag        string       `json:"tag"`
	Listen     string       `json:"listen"`
	ListenPort int          `json:"listen_port"`
	Users      []vlessUser  `json:"users"`
	TLS        *vlessTLS    `json:"tls,omitempty"`
}

type vlessUser struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
	Flow string `json:"flow,omitempty"`
}

type vlessTLS struct {
	Enabled         bool          `json:"enabled"`
	ServerName      string        `json:"server_name,omitempty"`
	CertificatePath string        `json:"certificate_path,omitempty"`
	KeyPath         string        `json:"key_path,omitempty"`
	Certificate     []string      `json:"certificate,omitempty"`
	Key             []string      `json:"key,omitempty"`
	Reality         *vlessReality `json:"reality,omitempty"`
}

type vlessReality struct {
	Enabled    bool             `json:"enabled"`
	Handshake  realityHandshake `json:"handshake"`
	PrivateKey string           `json:"private_key"`
	ShortID    []string         `json:"short_id,omitempty"`
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
	netSettings := nodeConfig.NetworkSettings
	if netSettings == nil {
		netSettings = make(map[string]interface{})
	}

	// Build users: use full UUID as name so Clash API / stats can attribute traffic
	vUsers := make([]vlessUser, len(users))
	for i, u := range users {
		vu := vlessUser{Name: u.UUID, UUID: u.UUID}
		if nodeConfig.Flow != "" && nodeConfig.Flow != "none" {
			vu.Flow = nodeConfig.Flow
		}
		vUsers[i] = vu
	}

	in := vlessInbound{
		Type:       "vless",
		Tag:        "vless-in",
		Listen:     listenAddr(nodeConfig.ListenIP),
		ListenPort: nodeConfig.ServerPort,
		Users:      vUsers,
	}

	// security=none / tls=0: plain VLESS without TLS/Reality
	if nodeConfig.TLS == 0 {
		return in, nil
	}

	// TLS or Reality enabled
	realitySettings, _ := tlsSettings["reality"].(map[string]interface{})
	privateKey := firstNonEmptyString(
		getMapString(realitySettings, "private_key"),
		getMapString(tlsSettings, "private_key", "reality_private_key"),
		getMapString(netSettings, "reality_private_key", "private_key"),
	)
	shortID := firstNonEmptyString(
		getMapString(realitySettings, "short_id"),
		getMapString(tlsSettings, "short_id", "reality_short_id"),
		getMapString(netSettings, "reality_short_id", "short_id"),
	)
	serverName := firstNonEmptyString(
		nodeConfig.ServerName,
		getMapString(tlsSettings, "server_name", "reality_server_name"),
		getMapString(netSettings, "reality_server_name", "server_name"),
	)

	// Prefer Reality only when private key is actually present
	privateKey = normalizeRealityKey(privateKey)
	if privateKey != "" || realitySettings != nil {
		if privateKey == "" {
			return vlessInbound{}, fmt.Errorf("security=reality but private_key is empty or invalid")
		}

		handshakeServer := firstNonEmptyString(serverName,
			getMapString(realitySettings, "server"),
		)
		handshakePort := 443
		if p := getMapInt(realitySettings, "server_port"); p > 0 {
			handshakePort = p
		}
		if hs, ok := realitySettings["handshake"].(map[string]interface{}); ok {
			if s := getMapString(hs, "server"); s != "" {
				handshakeServer = s
			}
			if p := getMapInt(hs, "server_port"); p > 0 {
				handshakePort = p
			}
		}
		if p := getMapInt(tlsSettings, "handshake_port", "reality_port"); p > 0 {
			handshakePort = p
		}
		if p := getMapInt(netSettings, "reality_port", "handshake_port"); p > 0 {
			handshakePort = p
		}
		if handshakeServer == "" {
			handshakeServer = "www.microsoft.com"
		}

		reality := &vlessReality{
			Enabled: true,
			Handshake: realityHandshake{
				Server:     handshakeServer,
				ServerPort: handshakePort,
			},
			PrivateKey: privateKey,
		}
		if shortID != "" {
			reality.ShortID = []string{shortID}
		}

		in.Tag = "vless-reality"
		in.TLS = &vlessTLS{
			Enabled:    true,
			ServerName: firstNonEmptyString(serverName, handshakeServer),
			Reality:    reality,
		}
		return in, nil
	}

	// Regular TLS
	tlsCfg := &vlessTLS{
		Enabled:    true,
		ServerName: serverName,
	}
	if nodeConfig.CertPEM != "" && nodeConfig.KeyPEM != "" {
		tlsCfg.Certificate = []string{nodeConfig.CertPEM}
		tlsCfg.Key = []string{nodeConfig.KeyPEM}
	}
	if certPath := getMapString(tlsSettings, "certificate_path", "cert_path"); certPath != "" {
		tlsCfg.CertificatePath = certPath
	}
	if keyPath := getMapString(tlsSettings, "key_path"); keyPath != "" {
		tlsCfg.KeyPath = keyPath
	}
	in.Tag = "vless-tls"
	in.TLS = tlsCfg
	return in, nil
}

func getMapString(m map[string]interface{}, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case string:
				if strings.TrimSpace(val) != "" {
					return strings.TrimSpace(val)
				}
			}
		}
	}
	return ""
}

func getMapInt(m map[string]interface{}, keys ...string) int {
	if m == nil {
		return 0
	}
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case float64:
				return int(val)
			case int:
				return val
			case int64:
				return int(val)
			case string:
				var n int
				if _, err := fmt.Sscanf(val, "%d", &n); err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func firstNonEmptyString(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// normalizeRealityKey accepts raw-url or std base64 private keys and returns raw-url form.
func normalizeRealityKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	// Try RawURL first (sing-box default), then StdEncoding with/without padding.
	decoders := []func(string) ([]byte, error){
		func(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) },
		func(s string) ([]byte, error) { return base64.URLEncoding.DecodeString(s) },
		func(s string) ([]byte, error) { return base64.RawStdEncoding.DecodeString(s) },
		func(s string) ([]byte, error) { return base64.StdEncoding.DecodeString(s) },
	}
	for _, decode := range decoders {
		if b, err := decode(key); err == nil && len(b) == 32 {
			return base64.RawURLEncoding.EncodeToString(b)
		}
	}
	return ""
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
	// Build users: name uses full UUID (hy2 password remains compact hex)
	hyUsers := make([]hysteria2User, len(users))
	for i, u := range users {
		pw := strings.ReplaceAll(u.UUID, "-", "")
		if len(pw) > 32 {
			pw = pw[:32]
		}
		hy := hysteria2User{Name: u.UUID, Password: pw}
		if u.SpeedLimit > 0 {
			hy.UpMbps = u.SpeedLimit
			hy.DownMbps = u.SpeedLimit
		}
		hyUsers[i] = hy
	}

	in := hysteria2Inbound{
		Type:       "hysteria2",
		Tag:        "hysteria2-in",
		Listen:     listenAddr(nodeConfig.ListenIP),
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
	// Build users: name = full UUID for consistent attribution
	tUsers := make([]tuicUser, len(users))
	for i, u := range users {
		tUsers[i] = tuicUser{
			Name:     u.UUID,
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
		Listen:            listenAddr(nodeConfig.ListenIP),
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
