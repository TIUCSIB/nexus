package service

import (
	"encoding/json"
	"fmt"
	"strings"

	nexusmodel "nexus/internal/model"
)

// ---------------------------------------------------------------------------
// sing-box top-level config structures
// ---------------------------------------------------------------------------

type singboxConfig struct {
	Log       logConfig   `json:"log"`
	Inbounds  []any       `json:"inbounds"`
	Outbounds []outbound  `json:"outbounds"`
	Route     routeConfig `json:"route"`
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
	Rules []routeRule `json:"rules"`
	Final string      `json:"final"`
}

type routeRule struct {
	Type     string   `json:"type"`
	Outbound string   `json:"outbound"`
	Protocol []string `json:"protocol,omitempty"`
}

// ---------------------------------------------------------------------------
// VLESS + Reality
// ---------------------------------------------------------------------------

type vlessConfigParams struct {
	ServerName      string `json:"server_name"`
	PrivateKey      string `json:"private_key"`
	ShortID         string `json:"short_id"`
	HandshakeServer string `json:"handshake_server"`
	HandshakePort   int    `json:"handshake_port"`
}

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

// ---------------------------------------------------------------------------
// Hysteria2
// ---------------------------------------------------------------------------

type hysteria2ConfigParams struct {
	UpMbps       int    `json:"up_mbps"`
	DownMbps     int    `json:"down_mbps"`
	ObfsPassword string `json:"obfs_password"`
	CertPath     string `json:"cert_path"`
	KeyPath      string `json:"key_path"`
}

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
}

type hysteria2TLS struct {
	Enabled         bool   `json:"enabled"`
	ServerName      string `json:"server_name,omitempty"`
	CertificatePath string `json:"certificate_path,omitempty"`
	KeyPath         string `json:"key_path,omitempty"`
}

type bandwidth struct {
	Up   string `json:"up"`
	Down string `json:"down"`
}

type obfsConfig struct {
	Type     string `json:"type"`
	Password string `json:"password"`
}

// ---------------------------------------------------------------------------
// TUIC
// ---------------------------------------------------------------------------

type tuicConfigParams struct {
	CongestionControl string `json:"congestion_control"`
	CertPath          string `json:"cert_path"`
	KeyPath           string `json:"key_path"`
}

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
	Enabled         bool   `json:"enabled"`
	ServerName      string `json:"server_name,omitempty"`
	CertificatePath string `json:"certificate_path,omitempty"`
	KeyPath         string `json:"key_path,omitempty"`
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// GenerateSingboxConfig produces a complete sing-box JSON configuration for
// the given node and active user set.
//
//   - ConfigMode == "auto"  : builds the config from scratch using ConfigJSON
//     protocol parameters and the node port.
//   - ConfigMode == "manual": treats ConfigJSON as a raw sing-box config and
//     only injects / replaces the users arrays.
func GenerateSingboxConfig(node nexusmodel.Node, users []nexusmodel.User) (string, error) {
	switch node.ConfigMode {
	case "manual":
		return generateManualConfig(node, users)
	default: // "auto"
		return generateAutoConfig(node, users)
	}
}

// ---------------------------------------------------------------------------
// Auto mode
// ---------------------------------------------------------------------------

func generateAutoConfig(node nexusmodel.Node, users []nexusmodel.User) (string, error) {
	cfg := baseConfig()

	switch strings.ToLower(node.Protocol) {
	case "vless":
		inbound, err := buildVLESSInbound(node, users)
		if err != nil {
			return "", err
		}
		cfg.Inbounds = append(cfg.Inbounds, inbound)

	case "hysteria2", "hy2":
		inbound, err := buildHysteria2Inbound(node, users)
		if err != nil {
			return "", err
		}
		cfg.Inbounds = append(cfg.Inbounds, inbound)

	case "tuic":
		inbound, err := buildTUICInbound(node, users)
		if err != nil {
			return "", err
		}
		cfg.Inbounds = append(cfg.Inbounds, inbound)

	default:
		return "", fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func buildVLESSInbound(node nexusmodel.Node, users []nexusmodel.User) (vlessInbound, error) {
	var params vlessConfigParams
	if err := json.Unmarshal([]byte(node.ConfigJSON), &params); err != nil {
		return vlessInbound{}, fmt.Errorf("parse VLESS config params: %w", err)
	}
	if params.HandshakePort == 0 {
		params.HandshakePort = 443
	}

	vUsers := make([]vlessUser, len(users))
	for i, u := range users {
		vUsers[i] = vlessUser{Name: u.Email, UUID: u.UUID}
	}

return vlessInbound{
			Type:       "vless",
			Tag:        "vless-reality",
			Listen:     "::",
			ListenPort: effectiveListenPort(node),
			Users:      vUsers,
		TLS: vlessTLS{
			Enabled:    true,
			ServerName: params.ServerName,
			Reality: vlessReality{
				Enabled: true,
				Handshake: realityHandshake{
					Server:     params.HandshakeServer,
					ServerPort: params.HandshakePort,
				},
				PrivateKey: params.PrivateKey,
				ShortID:    []string{params.ShortID},
			},
		},
	}, nil
}

func buildHysteria2Inbound(node nexusmodel.Node, users []nexusmodel.User) (hysteria2Inbound, error) {
	var params hysteria2ConfigParams
	if err := json.Unmarshal([]byte(node.ConfigJSON), &params); err != nil {
		return hysteria2Inbound{}, fmt.Errorf("parse Hysteria2 config params: %w", err)
	}

	certPath := params.CertPath
	if certPath == "" {
		certPath = "/etc/nexus/cert.pem"
	}
	keyPath := params.KeyPath
	if keyPath == "" {
		keyPath = "/etc/nexus/key.pem"
	}

	hyUsers := make([]hysteria2User, len(users))
	for i, u := range users {
		// Password = first 32 chars of UUID (dashes stripped).
		pw := strings.ReplaceAll(u.UUID, "-", "")
		if len(pw) > 32 {
			pw = pw[:32]
		}
		hyUsers[i] = hysteria2User{Name: u.Email, Password: pw}
	}

in := hysteria2Inbound{
			Type:       "hysteria2",
			Tag:        "hysteria2-in",
			Listen:     "::",
			ListenPort: effectiveListenPort(node),
		Users:      hyUsers,
		TLS: hysteria2TLS{
			Enabled:         true,
			CertificatePath: certPath,
			KeyPath:         keyPath,
		},
	}

	if params.UpMbps > 0 || params.DownMbps > 0 {
		in.Bandwidth = &bandwidth{
			Up:   fmt.Sprintf("%d mbps", params.UpMbps),
			Down: fmt.Sprintf("%d mbps", params.DownMbps),
		}
	}

	if params.ObfsPassword != "" {
		in.Obfs = &obfsConfig{
			Type:     "salamander",
			Password: params.ObfsPassword,
		}
	}

	return in, nil
}

func buildTUICInbound(node nexusmodel.Node, users []nexusmodel.User) (tuicInbound, error) {
	var params tuicConfigParams
	if err := json.Unmarshal([]byte(node.ConfigJSON), &params); err != nil {
		return tuicInbound{}, fmt.Errorf("parse TUIC config params: %w", err)
	}

	certPath := params.CertPath
	if certPath == "" {
		certPath = "/etc/nexus/cert.pem"
	}
	keyPath := params.KeyPath
	if keyPath == "" {
		keyPath = "/etc/nexus/key.pem"
	}
	congestion := params.CongestionControl
	if congestion == "" {
		congestion = "cubic"
	}

	tUsers := make([]tuicUser, len(users))
	for i, u := range users {
		tUsers[i] = tuicUser{
			Name:     u.Email,
			UUID:     u.UUID,
			Password: u.UUID,
		}
	}

return tuicInbound{
			Type:              "tuic",
			Tag:               "tuic-in",
			Listen:            "::",
			ListenPort:        effectiveListenPort(node),
		Users:             tUsers,
		CongestionControl: congestion,
		TLS: tuicTLS{
			Enabled:         true,
			CertificatePath: certPath,
			KeyPath:         keyPath,
		},
	}, nil
}

// ---------------------------------------------------------------------------
// Manual mode
// ---------------------------------------------------------------------------

func generateManualConfig(node nexusmodel.Node, users []nexusmodel.User) (string, error) {
	// Treat ConfigJSON as a full sing-box config, then inject users into
	// every inbound that has a users field.
	var raw map[string]any
	if err := json.Unmarshal([]byte(node.ConfigJSON), &raw); err != nil {
		return "", fmt.Errorf("parse manual config: %w", err)
	}

	inbounds, ok := raw["inbounds"].([]any)
	if !ok {
		out, _ := json.MarshalIndent(raw, "", "  ")
		return string(out), nil
	}

	for i, item := range inbounds {
		ib, ok := item.(map[string]any)
		if !ok {
			continue
		}
		protocol, _ := ib["type"].(string)
		ib["users"] = buildUsersForProtocol(protocol, users)
		inbounds[i] = ib
	}
	raw["inbounds"] = inbounds

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// effectiveListenPort returns the port the sing-box should actually listen on.
// If service_port is set, it takes priority; otherwise falls back to the
// connection port (port).
func effectiveListenPort(node nexusmodel.Node) int {
	if node.ServicePort > 0 {
		return node.ServicePort
	}
	return node.Port
}

func baseConfig() singboxConfig {
	return singboxConfig{
		Log: logConfig{
			Level:     "info",
			Timestamp: true,
		},
		Outbounds: []outbound{
			{Type: "direct", Tag: "direct"},
			{Type: "block", Tag: "block"},
		},
		Route: routeConfig{
			Rules: []routeRule{
				{Type: "logical", Outbound: "block", Protocol: []string{"dns"}},
			},
			Final: "direct",
		},
	}
}

// buildUsersForProtocol returns a []map[string]any suitable for injecting into
// a raw sing-box inbound config under the users key.
func buildUsersForProtocol(protocol string, users []nexusmodel.User) []map[string]any {
	result := make([]map[string]any, 0, len(users))
	switch strings.ToLower(protocol) {
	case "vless":
		for _, u := range users {
			result = append(result, map[string]any{
				"name": u.Email,
				"uuid": u.UUID,
			})
		}
	case "hysteria2", "hy2":
		for _, u := range users {
			pw := strings.ReplaceAll(u.UUID, "-", "")
			if len(pw) > 32 {
				pw = pw[:32]
			}
			result = append(result, map[string]any{
				"name":     u.Email,
				"password": pw,
			})
		}
	case "tuic":
		for _, u := range users {
			result = append(result, map[string]any{
				"name":     u.Email,
				"uuid":     u.UUID,
				"password": u.UUID,
			})
		}
	default:
		for _, u := range users {
			result = append(result, map[string]any{
				"name": u.Email,
				"uuid": u.UUID,
			})
		}
	}
	return result
}
