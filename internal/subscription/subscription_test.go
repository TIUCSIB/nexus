package subscription

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"nexus/internal/model"

	"gopkg.in/yaml.v3"
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

func findClashProxyByName(t *testing.T, body []byte, name string) map[string]interface{} {
	t.Helper()

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		t.Fatalf("failed to parse clash yaml: %v", err)
	}

	proxies, ok := cfg["proxies"].([]interface{})
	if !ok {
		t.Fatalf("clash config proxies should be array")
	}

	for _, item := range proxies {
		proxy, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if proxyName, _ := proxy["name"].(string); proxyName == name {
			return proxy
		}
	}

	t.Fatalf("proxy %s not found", name)
	return nil
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
	node := makeNode(1, "hysteria2")
	node.NetworkSettings = `{"tls_server_name":"node.now.cc","tls_allow_insecure":true,"bandwidth_up":300,"bandwidth_down":500,"obfs_open":true,"obfs_type":"salamander","obfs_password":"secret-pass"}`
	out, err := GenerateSingbox([]model.Node{node}, makeUser())
	if err != nil {
		t.Fatalf("GenerateSingbox Hysteria2 error: %v", err)
	}
	body := string(out)
	if !strings.Contains(body, "hysteria2") {
		t.Error("output should contain hysteria2 outbound")
	}
	if !strings.Contains(body, `"up_mbps": 300`) {
		t.Error("output should contain normalized up_mbps")
	}
	if !strings.Contains(body, `"down_mbps": 500`) {
		t.Error("output should contain normalized down_mbps")
	}
	if !strings.Contains(body, `"insecure": true`) {
		t.Error("output should contain insecure tls flag")
	}
}

func TestGenerateSingbox_TUIC(t *testing.T) {
	node := makeNode(1, "tuic")
	node.NetworkSettings = `{"tls_server_name":"tuic.example.com","tls_allow_insecure":true,"congestion_control":"bbr"}`
	out, err := GenerateSingbox([]model.Node{node}, makeUser())
	if err != nil {
		t.Fatalf("GenerateSingbox TUIC error: %v", err)
	}
	body := string(out)
	if !strings.Contains(body, "tuic") {
		t.Error("output should contain tuic outbound")
	}
	if !strings.Contains(body, `"server_name": "tuic.example.com"`) {
		t.Error("output should contain normalized tuic server_name")
	}
	if !strings.Contains(body, `"insecure": true`) {
		t.Error("output should contain insecure tls flag")
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
	node := makeNode(1, "hysteria2")
	node.Address = "127.0.0.1"
	node.Port = 58943
	node.NetworkSettings = `{"tls_server_name":"node.now.cc","tls_allow_insecure":true,"bandwidth_up":300,"bandwidth_down":500,"obfs_open":true,"obfs_type":"salamander","obfs_password":"0iaDOtsbl0w0M84r"}`
	out, err := GenerateClash([]model.Node{node}, makeUser(), "Nexus")
	if err != nil {
		t.Fatalf("GenerateClash Hysteria2 error: %v", err)
	}

	proxy := findClashProxyByName(t, out, node.Name)
	if proxy["type"] != "hysteria2" {
		t.Fatalf("expected hysteria2 proxy, got %v", proxy["type"])
	}
	if proxy["obfs"] != "salamander" {
		t.Fatalf("expected obfs salamander, got %#v", proxy["obfs"])
	}
	if proxy["obfs-password"] != "0iaDOtsbl0w0M84r" {
		t.Fatalf("expected obfs-password to be exported, got %#v", proxy["obfs-password"])
	}
	if _, ok := proxy["obfs"].(map[string]interface{}); ok {
		t.Fatal("clash hysteria2 obfs should not be exported as object")
	}
	if proxy["sni"] != "node.now.cc" {
		t.Fatalf("expected sni node.now.cc, got %#v", proxy["sni"])
	}
	if proxy["skip-cert-verify"] != true {
		t.Fatalf("expected skip-cert-verify true, got %#v", proxy["skip-cert-verify"])
	}
	if proxy["up"] != 300 {
		t.Fatalf("expected up 300, got %#v", proxy["up"])
	}
	if proxy["down"] != 500 {
		t.Fatalf("expected down 500, got %#v", proxy["down"])
	}
}

func TestGenerateClash_VLESSSecurityModes(t *testing.T) {
	user := makeUser()

	noneNode := makeNode(1, "vless")
	noneNode.Name = "vless-none"
	noneNode.Security = "none"
	noneNode.FlowControl = "none"
	noneNode.NetworkSettings = `{"server_name":"ignored.example.com"}`

	tlsNode := makeNode(2, "vless")
	tlsNode.Name = "vless-tls"
	tlsNode.Security = "tls"
	tlsNode.FlowControl = "none"
	tlsNode.NetworkSettings = `{"server_name":"tls.example.com","allow_insecure":true}`

	realityNode := makeNode(3, "vless")
	realityNode.Name = "vless-reality"
	realityNode.Security = "reality"
	realityNode.FlowControl = "xtls-rprx-vision"
	realityNode.NetworkSettings = `{"reality_server_name":"reality.example.com","reality_public_key":"pub-key","reality_short_id":"abcd","allow_insecure":true}`

	out, err := GenerateClash([]model.Node{noneNode, tlsNode, realityNode}, user, "Nexus")
	if err != nil {
		t.Fatalf("GenerateClash VLESS security modes error: %v", err)
	}

	noneProxy := findClashProxyByName(t, out, noneNode.Name)
	if noneProxy["tls"] != false {
		t.Fatalf("security=none should export tls=false, got %#v", noneProxy["tls"])
	}
	if _, ok := noneProxy["flow"]; ok {
		t.Fatal("flow should not be exported when flow_control=none")
	}
	if _, ok := noneProxy["servername"]; ok {
		t.Fatal("security=none should not export servername")
	}

	tlsProxy := findClashProxyByName(t, out, tlsNode.Name)
	if tlsProxy["tls"] != true {
		t.Fatalf("security=tls should export tls=true, got %#v", tlsProxy["tls"])
	}
	if tlsProxy["servername"] != "tls.example.com" {
		t.Fatalf("security=tls should export servername, got %#v", tlsProxy["servername"])
	}
	if tlsProxy["skip-cert-verify"] != true {
		t.Fatalf("security=tls should export skip-cert-verify, got %#v", tlsProxy["skip-cert-verify"])
	}
	if _, ok := tlsProxy["reality-opts"]; ok {
		t.Fatal("security=tls should not export reality-opts")
	}

	realityProxy := findClashProxyByName(t, out, realityNode.Name)
	if realityProxy["tls"] != true {
		t.Fatalf("security=reality should export tls=true, got %#v", realityProxy["tls"])
	}
	if realityProxy["flow"] != "xtls-rprx-vision" {
		t.Fatalf("security=reality should export configured flow, got %#v", realityProxy["flow"])
	}
	realityOpts, ok := realityProxy["reality-opts"].(map[string]interface{})
	if !ok {
		t.Fatal("security=reality should export reality-opts")
	}
	if realityOpts["public-key"] != "pub-key" {
		t.Fatalf("expected reality public key, got %#v", realityOpts["public-key"])
	}
	if realityOpts["short-id"] != "abcd" {
		t.Fatalf("expected reality short-id, got %#v", realityOpts["short-id"])
	}
}

func TestGenerateClashMeta_VLESS(t *testing.T) {
	nodes := []model.Node{makeNode(1, "vless")}
	out, err := GenerateClashMeta(nodes, makeUser(), "Nexus")
	if err != nil {
		t.Fatalf("GenerateClashMeta VLESS error: %v", err)
	}
	if !strings.Contains(string(out), "vless") {
		t.Error("clashmeta output should contain vless proxy")
	}
}

func TestGenerateStash_VLESS(t *testing.T) {
	nodes := []model.Node{makeNode(1, "vless")}
	out, err := GenerateStash(nodes, makeUser(), "Nexus")
	if err != nil {
		t.Fatalf("GenerateStash VLESS error: %v", err)
	}
	if !strings.Contains(string(out), "vless") {
		t.Error("stash output should contain vless proxy")
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

func TestBuildHysteria2URI(t *testing.T) {
	node := makeNode(1, "hysteria2")
	node.NetworkSettings = `{"tls_server_name":"node.now.cc","tls_allow_insecure":true,"bandwidth_up":300,"bandwidth_down":500,"obfs_open":true,"obfs_type":"salamander","obfs_password":"secret-pass"}`
	uri := buildHysteria2URI(node, makeUser(), ParseNodeParams(node.ConfigJSON, node.NetworkSettings))
	if !strings.Contains(uri, "obfs=salamander") {
		t.Fatalf("hysteria2 uri should contain obfs type, got %s", uri)
	}
	if !strings.Contains(uri, "obfs-password=secret-pass") {
		t.Fatalf("hysteria2 uri should contain obfs password, got %s", uri)
	}
	if !strings.Contains(uri, "upmbps=300") {
		t.Fatalf("hysteria2 uri should contain upmbps, got %s", uri)
	}
	if !strings.Contains(uri, "downmbps=500") {
		t.Fatalf("hysteria2 uri should contain downmbps, got %s", uri)
	}
}

// --- Surge ---

func TestGenerateSurge(t *testing.T) {
	nodes := []model.Node{makeNode(1, "vless")}
	out, err := GenerateSurge(nodes, makeUser())
	if err != nil {
		t.Fatalf("GenerateSurge error: %v", err)
	}
	body := string(out)
	if !strings.Contains(body, "[Proxy]") {
		t.Error("surge config should contain [Proxy] section")
	}
	if !strings.Contains(body, "DOMAIN-SUFFIX,services.googleapis.cn,Proxy") {
		t.Error("surge config should contain extended default rules")
	}
	if !strings.Contains(body, "REJECT-TINYGIF") {
		t.Error("surge config should contain ad-block rules")
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

func TestParseNodeParams_Aliases(t *testing.T) {
	params := ParseNodeParams(
		`{"obfs":{"open":true,"type":"salamander","password":"config-pass"},"bandwidth":{"up":50,"down":60}}`,
		`{"tls_server_name":"node.now.cc","bandwidth_up":300,"bandwidth_down":500,"obfs_open":true,"obfs_type":"salamander","obfs_password":"secret-pass","tls_allow_insecure":true,"congestion_control":"bbr","reality_server_name":"reality.example.com","reality_port":8443,"reality_public_key":"pub-key","reality_short_id":"abcd"}`,
	)
	if params.ServerName != "node.now.cc" {
		t.Fatalf("expected tls_server_name alias, got %s", params.ServerName)
	}
	if params.HandshakeHost != "reality.example.com" {
		t.Fatalf("expected reality_server_name alias, got %s", params.HandshakeHost)
	}
	if params.HandshakePort != 8443 {
		t.Fatalf("expected reality_port alias, got %d", params.HandshakePort)
	}
	if params.PublicKey != "pub-key" {
		t.Fatalf("expected reality_public_key alias, got %s", params.PublicKey)
	}
	if params.ShortID != "abcd" {
		t.Fatalf("expected reality_short_id alias, got %s", params.ShortID)
	}
	if params.UpMbps != 300 || params.DownMbps != 500 {
		t.Fatalf("expected bandwidth aliases 300/500, got %d/%d", params.UpMbps, params.DownMbps)
	}
	if !params.ObfsEnabled {
		t.Fatal("expected obfs_open alias to enable obfs")
	}
	if params.ObfsType != "salamander" {
		t.Fatalf("expected obfs_type alias, got %s", params.ObfsType)
	}
	if params.ObfsPassword != "secret-pass" {
		t.Fatalf("expected obfs_password alias, got %s", params.ObfsPassword)
	}
	if !params.AllowInsecure {
		t.Fatal("expected tls_allow_insecure alias to be true")
	}
	if params.CongestionCtrl != "bbr" {
		t.Fatalf("expected congestion_control alias, got %s", params.CongestionCtrl)
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

func TestValidateSubscriptionTemplate(t *testing.T) {
	if err := ValidateSubscriptionTemplate(SettingSubscribeTemplateSingbox, defaultSingboxTemplate()); err != nil {
		t.Fatalf("default singbox template should be valid: %v", err)
	}
	if err := ValidateSubscriptionTemplate(SettingSubscribeTemplateClash, defaultClashTemplate()); err != nil {
		t.Fatalf("default clash template should be valid: %v", err)
	}
	if err := ValidateSubscriptionTemplate(SettingSubscribeTemplateSurge, defaultSurgeTemplate()); err != nil {
		t.Fatalf("default surge template should be valid: %v", err)
	}
}

func TestValidateSubscriptionTemplate_InvalidSurgeTemplate(t *testing.T) {
	err := ValidateSubscriptionTemplate(SettingSubscribeTemplateSurge, "[Proxy]\n$proxies\n")
	if err == nil {
		t.Fatal("invalid surge template should return error")
	}
}

func TestGetDefaultSubscriptionTemplate(t *testing.T) {
	if GetDefaultSubscriptionTemplate(SettingSubscribeTemplateClash) == "" {
		t.Fatal("default clash template should not be empty")
	}
	if GetDefaultSubscriptionTemplate(SettingSubscribeTemplateSingbox) == "" {
		t.Fatal("default singbox template should not be empty")
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
