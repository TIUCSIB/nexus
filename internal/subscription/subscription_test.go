package subscription

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"nexus/internal/model"
)

func makeNode(id uint, protocol string) model.Node {
	now := time.Now()
	return model.Node{
		ID:              id,
		Name:            "test-" + protocol,
		Address:         "1.2.3.4",
		Port:            443,
		Protocol:        protocol,
		Security:        "reality",
		FlowControl:     "xtls-rprx-vision",
		ConfigJSON:      `{"server_name":"test.example.com","public_key":"test-public-key","short_id":"1234","handshake_host":"hs.example.com","handshake_port":443}`,
		NetworkSettings: "",
		Status:          1,
		Sort:            100,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func makeUser() model.User {
	return model.User{
		ID:     1,
		UUID:   "550e8400-e29b-41d4-a716-446655440000",
		Email:  "test@example.com",
		Status: 1,
	}
}

// --- Singbox ---

func TestGenerateSingbox_EmptyNodes(t *testing.T) {
	out, err := GenerateSingbox([]model.Node{}, makeUser())
	if err != nil {
		t.Fatalf("GenerateSingbox empty nodes returned error: %v", err)
	}
	if !strings.Contains(string(out), "direct") {
		t.Error("empty node config should contain direct outbound")
	}
}

func TestGenerateSingbox_VLESS(t *testing.T) {
	nodes := []model.Node{makeNode(1, "vless")}
	out, err := GenerateSingbox(nodes, makeUser())
	if err != nil {
		t.Fatalf("GenerateSingbox VLESS error: %v", err)
	}

	var cfg map[string]interface{}
	json.Unmarshal(out, &cfg)

	if cfg["dns"] == nil {
		t.Error("singbox config should contain dns section")
	}
}

func TestGenerateSingbox_Hysteria2(t *testing.T) {
	nodes := []model.Node{makeNode(1, "hysteria2")}
	out, err := GenerateSingbox(nodes, makeUser())
	if err != nil {
		t.Fatalf("GenerateSingbox Hysteria2 error: %v", err)
	}
	if !strings.Contains(string(out), "hysteria2") {
		t.Error("output should contain hysteria2 outbound")
	}
}

func TestGenerateSingbox_TUIC(t *testing.T) {
	nodes := []model.Node{makeNode(1, "tuic")}
	out, err := GenerateSingbox(nodes, makeUser())
	if err != nil {
		t.Fatalf("GenerateSingbox TUIC error: %v", err)
	}
	if !strings.Contains(string(out), "tuic") {
		t.Error("output should contain tuic outbound")
	}
}

func TestGenerateSingbox_MultiProtocol(t *testing.T) {
	nodes := []model.Node{
		makeNode(1, "vless"),
		makeNode(2, "hysteria2"),
		makeNode(3, "tuic"),
	}
	out, err := GenerateSingbox(nodes, makeUser())
	if err != nil {
		t.Fatalf("GenerateSingbox multi error: %v", err)
	}
	body := string(out)
	for _, proto := range []string{"vless", "hysteria2", "tuic"} {
		if !strings.Contains(body, proto) {
			t.Errorf("missing protocol %s in output", proto)
		}
	}
}

// --- Clash ---

func TestGenerateClash_EmptyNodes(t *testing.T) {
	out, err := GenerateClash([]model.Node{}, makeUser(), "Nexus")
	if err != nil {
		t.Fatalf("GenerateClash empty nodes error: %v", err)
	}
	if !strings.Contains(string(out), "proxies:") {
		t.Error("empty clash config should contain proxies section")
	}
}

func TestGenerateClash_VLESS(t *testing.T) {
	nodes := []model.Node{makeNode(1, "vless")}
	out, err := GenerateClash(nodes, makeUser(), "Nexus")
	if err != nil {
		t.Fatalf("GenerateClash VLESS error: %v", err)
	}
	if !strings.Contains(string(out), "vless") {
		t.Error("clash output should contain vless proxy")
	}
	if !strings.Contains(string(out), "fake-ip") {
		t.Error("clash output should contain fake-ip dns")
	}
}

func TestGenerateClash_Hysteria2(t *testing.T) {
	nodes := []model.Node{makeNode(1, "hysteria2")}
	out, err := GenerateClash(nodes, makeUser(), "Nexus")
	if err != nil {
		t.Fatalf("GenerateClash Hysteria2 error: %v", err)
	}
	if !strings.Contains(string(out), "hysteria2") {
		t.Error("clash output should contain hysteria2 proxy")
	}
}

// --- V2RayN / URI ---

func TestGenerateV2RayN(t *testing.T) {
	nodes := []model.Node{makeNode(1, "vless")}
	out, err := GenerateV2RayN(nodes, makeUser())
	if err != nil {
		t.Fatalf("GenerateV2RayN error: %v", err)
	}
	if len(out) == 0 {
		t.Error("v2rayn output should not be empty")
	}
}

func TestGenerateV2RayN_DisabledNode(t *testing.T) {
	node := makeNode(1, "vless")
	node.Status = 0 // disabled node
	nodes := []model.Node{node}
	out, err := GenerateV2RayN(nodes, makeUser())
	if err != nil {
		t.Fatalf("GenerateV2RayN disabled node error: %v", err)
	}
	if len(out) != 0 {
		t.Error("v2rayn should return empty for disabled nodes only")
	}
}

// --- Surge ---

func TestGenerateSurge(t *testing.T) {
	nodes := []model.Node{makeNode(1, "vless")}
	out, err := GenerateSurge(nodes, makeUser())
	if err != nil {
		t.Fatalf("GenerateSurge error: %v", err)
	}
	if !strings.Contains(string(out), "[Proxy]") {
		t.Error("surge config should contain [Proxy] section")
	}
}

// --- ParseNodeParams ---

func TestParseNodeParams_JSON(t *testing.T) {
	params := ParseNodeParams(`{"server_name":"my.example.com","public_key":"pk123"}`, "")
	if params.ServerName != "my.example.com" {
		t.Errorf("expected my.example.com, got %s", params.ServerName)
	}
	if params.PublicKey != "pk123" {
		t.Errorf("expected pk123, got %s", params.PublicKey)
	}
}

func TestParseNodeParams_NetworkSettingsOverride(t *testing.T) {
	params := ParseNodeParams(
		`{"server_name":"old.example.com","public_key":"pk123"}`,
		`{"server_name":"new.example.com"}`,
	)
	if params.ServerName != "new.example.com" {
		t.Errorf("expected network settings to override: new.example.com, got %s", params.ServerName)
	}
	if params.PublicKey != "pk123" {
		t.Errorf("public_key should still come from config json: pk123, got %s", params.PublicKey)
	}
}

func TestParseNodeParams_Empty(t *testing.T) {
	params := ParseNodeParams("", "")
	if params.ServerName != "" {
		t.Error("empty config should return empty params")
	}
}

// --- InfoNode ---

func TestGetInfoNodeNames(t *testing.T) {
	user := makeUser()
	user.ExpiredAt = &[]time.Time{time.Now().Add(30 * 24 * time.Hour)}[0]
	user.TrafficLimit = 100 * 1024 * 1024 * 1024 // 100 GB
	user.TrafficUsed = 30 * 1024 * 1024 * 1024   // 30 GB

	expiryName, trafficName := GetInfoNodeNames(user)
	if expiryName == "" {
		t.Error("expiry name should not be empty")
	}
	if trafficName == "" {
		t.Error("traffic name should not be empty")
	}
	if !strings.Contains(trafficName, "70") {
		t.Errorf("expected 70 GB remaining, got %s", trafficName)
	}
}

func TestGetInfoNodeNames_NoLimit(t *testing.T) {
	user := makeUser()
	user.TrafficLimit = 0 // unlimited

	_, trafficName := GetInfoNodeNames(user)
	if !strings.Contains(trafficName, "无限") {
		t.Errorf("unlimited traffic should show 无限, got %s", trafficName)
	}
}

func TestGetInfoNodeNames_Expired(t *testing.T) {
	user := makeUser()
	user.ExpiredAt = &[]time.Time{time.Now().Add(-24 * time.Hour)}[0]

	expiryName, _ := GetInfoNodeNames(user)
	if !strings.Contains(expiryName, "过期") {
		t.Errorf("expired user should show 已过期, got %s", expiryName)
	}
}

// --- ParseNodeParams via URIs (integration-style) ---

func TestBuildVlessURI(t *testing.T) {
	node := makeNode(1, "vless")
	user := makeUser()
	params := ParseNodeParams(node.ConfigJSON, node.NetworkSettings)

	uri := buildVlessURI(node, user, params)
	if !strings.HasPrefix(uri, "vless://") {
		t.Errorf("vless URI should start with vless://, got %s", uri)
	}
	if !strings.Contains(uri, user.UUID) {
		t.Errorf("vless URI should contain user UUID, got %s", uri)
	}
}