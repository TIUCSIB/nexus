package kernel

import (
	"strings"
	"testing"
)

func makeNodeConfig(protocol string) NodeConfig {
	return NodeConfig{
		Protocol:   protocol,
		ListenIP:   "0.0.0.0",
		ServerPort: 443,
		StatsPort:  9090,
		Network:    "tcp",
		TLS:        1,
		ServerName: "test.example.com",
		TLSSettings: map[string]interface{}{
			"reality": map[string]interface{}{
				"private_key": "cAaWAHKAOQwFkSKUmDlyhKHyzj2DIAsfqUr-OZ3v8Gs",
				"short_id":    "1234",
				"handshake": map[string]interface{}{
					"server":      "hs.example.com",
					"server_port": 443,
				},
			},
			"server_name": "test.example.com",
		},
	}
}

func makeUsers() []User {
	return []User{
		{ID: 1, UUID: "550e8400-e29b-41d4-a716-446655440000", SpeedLimit: 100, DeviceLimit: 0},
		{ID: 2, UUID: "660e8400-e29b-41d4-a716-446655440001", SpeedLimit: 0, DeviceLimit: 3},
	}
}

func TestGenerateSingboxConfig_VLESS(t *testing.T) {
	cfg := makeNodeConfig("vless")
	out, err := GenerateSingboxConfig(cfg, makeUsers())
	if err != nil {
		t.Fatalf("GenerateSingboxConfig VLESS error: %v", err)
	}
	if !strings.Contains(out, "vless") {
		t.Error("output should contain vless inbound")
	}
	if !strings.Contains(out, "dns") {
		t.Error("output should contain dns section")
	}
	if !strings.Contains(out, "local") {
		t.Error("output should contain local server")
	}
	if !strings.Contains(out, "direct") {
		t.Error("output should contain direct outbound")
	}
}

func TestGenerateSingboxConfig_Hysteria2(t *testing.T) {
	cfg := makeNodeConfig("hysteria2")
	cfg.UpMbps = 100
	cfg.DownMbps = 500
	cfg.ObfsPassword = "test-obfs"

	out, err := GenerateSingboxConfig(cfg, makeUsers())
	if err != nil {
		t.Fatalf("GenerateSingboxConfig Hysteria2 error: %v", err)
	}
	if !strings.Contains(out, "hysteria2") {
		t.Error("output should contain hysteria2 inbound")
	}
	if !strings.Contains(out, "bandwidth") {
		t.Error("hysteria2 output should contain bandwidth")
	}
	if !strings.Contains(out, "obfs") {
		t.Error("hysteria2 output should contain obfs")
	}
}

func TestGenerateSingboxConfig_TUIC(t *testing.T) {
	cfg := makeNodeConfig("tuic")
	cfg.CongestionControl = "bbr"

	out, err := GenerateSingboxConfig(cfg, makeUsers())
	if err != nil {
		t.Fatalf("GenerateSingboxConfig TUIC error: %v", err)
	}
	if !strings.Contains(out, "tuic") {
		t.Error("output should contain tuic inbound")
	}
	if !strings.Contains(out, "bbr") {
		t.Error("tuic output should contain congestion_control")
	}
}

func TestGenerateSingboxConfig_UnsupportedProtocol(t *testing.T) {
	cfg := makeNodeConfig("vmess")
	_, err := GenerateSingboxConfig(cfg, makeUsers())
	if err == nil {
		t.Error("unsupported protocol should return error")
	}
	if !strings.Contains(err.Error(), "unsupported protocol") {
		t.Errorf("expected unsupported protocol error, got: %v", err)
	}
}

func TestGenerateSingboxConfig_EmptyUsers(t *testing.T) {
	cfg := makeNodeConfig("vless")
	out, err := GenerateSingboxConfig(cfg, []User{})
	if err != nil {
		t.Fatalf("GenerateSingboxConfig empty users error: %v", err)
	}
	if !strings.Contains(out, "direct") {
		t.Error("output should contain standard outbounds even with no users")
	}
}

func TestGenerateSingboxConfig_Routes(t *testing.T) {
	cfg := makeNodeConfig("vless")
	cfg.Routes = []RouteRule{
		{
			Match:       []string{"domain:google.com", "domain:youtube.com"},
			Action:      "route",
			ActionValue: "proxy",
		},
		{
			Match:  []string{"ip:10.0.0.0/8"},
			Action: "direct",
		},
	}

	out, err := GenerateSingboxConfig(cfg, makeUsers())
	if err != nil {
		t.Fatalf("GenerateSingboxConfig with routes error: %v", err)
	}
	if !strings.Contains(out, "google.com") {
		t.Error("route rules should include google.com")
	}
}

func TestGenerateSingboxConfig_CustomOutbounds(t *testing.T) {
	cfg := makeNodeConfig("vless")
	cfg.CustomOutbounds = []CustomOutbound{
		{
			Tag:      "proxy-warp",
			Protocol: "wireguard",
		},
	}

	out, err := GenerateSingboxConfig(cfg, makeUsers())
	if err != nil {
		t.Fatalf("GenerateSingboxConfig with custom outbounds error: %v", err)
	}
	if !strings.Contains(out, "proxy-warp") {
		t.Error("custom outbounds should be included in output")
	}
}

func TestGenerateSingboxConfig_Hysteria2NoObfs(t *testing.T) {
	cfg := makeNodeConfig("hysteria2")
	cfg.ObfsPassword = ""

	out, err := GenerateSingboxConfig(cfg, makeUsers())
	if err != nil {
		t.Fatalf("GenerateSingboxConfig Hysteria2 no obfs error: %v", err)
	}
	if strings.Contains(out, "obfs") {
		t.Error("hysteria2 without obfs should not contain obfs in config")
	}
}

func TestGenerateSingboxConfig_VLESS_None(t *testing.T) {
	cfg := makeNodeConfig("vless")
	cfg.TLS = 0
	cfg.TLSSettings = nil
	out, err := GenerateSingboxConfig(cfg, makeUsers())
	if err != nil {
		t.Fatalf("GenerateSingboxConfig VLESS none error: %v", err)
	}
	if strings.Contains(out, "reality") {
		t.Error("security=none should not contain reality")
	}
	if strings.Contains(out, `"tls"`) {
		// bare VLESS may still have no tls block; if present must not enable reality
		if strings.Contains(out, `"enabled": true`) && strings.Contains(out, "private_key") {
			t.Error("security=none should not enable reality TLS")
		}
	}
	if !strings.Contains(out, "vless") {
		t.Error("output should contain vless inbound")
	}
}
