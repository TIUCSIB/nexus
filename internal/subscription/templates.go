package subscription

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"nexus/internal/database"

	"gopkg.in/yaml.v3"
)

const (
	SettingSubscribeTemplateSingbox   = "subscribe_template_singbox"
	SettingSubscribeTemplateClash     = "subscribe_template_clash"
	SettingSubscribeTemplateClashMeta = "subscribe_template_clashmeta"
	SettingSubscribeTemplateStash     = "subscribe_template_stash"
	SettingSubscribeTemplateSurge     = "subscribe_template_surge"
	SettingSubscribeTemplateSurfboard = "subscribe_template_surfboard"

	clashAutoProxyPlaceholder       = "$auto_proxy"
	singboxAutoOutboundsPlaceholder = "$auto_outbounds"
	templateAppNamePlaceholder      = "$app_name"
	templateProxiesPlaceholder      = "$proxies"
	templateProxyGroupPlaceholder   = "$proxy_group"
	templateRulesPlaceholder        = "$rules"
)

var subscriptionTemplateKeys = []string{
	SettingSubscribeTemplateSingbox,
	SettingSubscribeTemplateClash,
	SettingSubscribeTemplateClashMeta,
	SettingSubscribeTemplateStash,
	SettingSubscribeTemplateSurge,
	SettingSubscribeTemplateSurfboard,
}

func SubscriptionTemplateKeys() []string {
	keys := make([]string, len(subscriptionTemplateKeys))
	copy(keys, subscriptionTemplateKeys)
	return keys
}

func GetSubscriptionTemplate(settingKey string) string {
	if value := strings.TrimSpace(database.GetSetting(settingKey)); value != "" {
		return value
	}
	return GetDefaultSubscriptionTemplate(settingKey)
}

func GetDefaultSubscriptionTemplates() map[string]string {
	return map[string]string{
		SettingSubscribeTemplateSingbox:   GetDefaultSubscriptionTemplate(SettingSubscribeTemplateSingbox),
		SettingSubscribeTemplateClash:     GetDefaultSubscriptionTemplate(SettingSubscribeTemplateClash),
		SettingSubscribeTemplateClashMeta: GetDefaultSubscriptionTemplate(SettingSubscribeTemplateClashMeta),
		SettingSubscribeTemplateStash:     GetDefaultSubscriptionTemplate(SettingSubscribeTemplateStash),
		SettingSubscribeTemplateSurge:     GetDefaultSubscriptionTemplate(SettingSubscribeTemplateSurge),
		SettingSubscribeTemplateSurfboard: GetDefaultSubscriptionTemplate(SettingSubscribeTemplateSurfboard),
	}
}

func GetDefaultSubscriptionTemplate(settingKey string) string {
	switch settingKey {
	case SettingSubscribeTemplateSingbox:
		return defaultSingboxTemplate()
	case SettingSubscribeTemplateClash, SettingSubscribeTemplateClashMeta, SettingSubscribeTemplateStash:
		return defaultClashTemplate()
	case SettingSubscribeTemplateSurge:
		return defaultSurgeTemplate()
	case SettingSubscribeTemplateSurfboard:
		return defaultSurfboardTemplate()
	default:
		return ""
	}
}

func GetSubscriptionAppName() string {
	return database.GetSettingDefault("app_name", "Proxy")
}

func ValidateSubscriptionTemplate(settingKey, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	switch settingKey {
	case SettingSubscribeTemplateSingbox:
		return validateSingboxTemplate(trimmed)
	case SettingSubscribeTemplateClash, SettingSubscribeTemplateClashMeta, SettingSubscribeTemplateStash:
		return validateClashTemplate(trimmed)
	case SettingSubscribeTemplateSurge, SettingSubscribeTemplateSurfboard:
		return validateTextSubscriptionTemplate(trimmed)
	default:
		return nil
	}
}

func validateClashTemplate(content string) error {
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return fmt.Errorf("Clash 模板 YAML 格式无效: %w", err)
	}
	if _, ok := data["proxies"]; !ok {
		return errors.New("Clash 模板缺少 proxies 字段")
	}
	if _, ok := data["proxy-groups"]; !ok {
		return errors.New("Clash 模板缺少 proxy-groups 字段")
	}
	if _, ok := data["rules"]; !ok {
		return errors.New("Clash 模板缺少 rules 字段")
	}
	return nil
}

func validateSingboxTemplate(content string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return fmt.Errorf("Sing-box 模板 JSON 格式无效: %w", err)
	}
	if _, ok := data["outbounds"]; !ok {
		return errors.New("Sing-box 模板缺少 outbounds 字段")
	}
	if _, ok := data["route"]; !ok {
		return errors.New("Sing-box 模板缺少 route 字段")
	}
	return nil
}

func validateTextSubscriptionTemplate(content string) error {
	for _, placeholder := range []string{templateProxiesPlaceholder, templateProxyGroupPlaceholder, templateRulesPlaceholder} {
		if !strings.Contains(content, placeholder) {
			return fmt.Errorf("模板缺少占位符 %s", placeholder)
		}
	}
	return nil
}

func defaultClashTemplate() string {
	cfg := clashConfig{
		MixedPort:          7890,
		AllowLan:           true,
		BindAddress:        "*",
		Mode:               "rule",
		LogLevel:           "info",
		ExternalController: "127.0.0.1:9090",
		UnifiedDelay:       true,
		TCPConcurrent:      true,
		DNS: &clashDNS{
			Enable:            true,
			IPv6:              false,
			DefaultNameserver: []string{"223.5.5.5", "119.29.29.29"},
			EnhancedMode:      "fake-ip",
			FakeIPRange:       "198.18.0.1/16",
			UseHosts:          true,
			NameserverPolicy: map[string]string{
				"+.google.com":            "https://dns.cloudflare.com/dns-query",
				"+.googleapis.com":        "https://dns.cloudflare.com/dns-query",
				"+.googleapis.cn":         "https://dns.cloudflare.com/dns-query",
				"+.googlevideo.com":       "https://dns.cloudflare.com/dns-query",
				"+.gstatic.com":           "https://dns.cloudflare.com/dns-query",
				"+.youtube.com":           "https://dns.cloudflare.com/dns-query",
				"+.youtu.be":              "https://dns.cloudflare.com/dns-query",
				"+.facebook.com":          "https://dns.cloudflare.com/dns-query",
				"+.twitter.com":           "https://dns.cloudflare.com/dns-query",
				"+.x.com":                 "https://dns.cloudflare.com/dns-query",
				"+.github.com":            "https://dns.cloudflare.com/dns-query",
				"+.githubusercontent.com": "https://dns.cloudflare.com/dns-query",
				"+.openai.com":            "https://dns.cloudflare.com/dns-query",
				"+.chatgpt.com":           "https://dns.cloudflare.com/dns-query",
				"+.anthropic.com":         "https://dns.cloudflare.com/dns-query",
			},
			Nameserver: []string{
				"https://doh.pub/dns-query",
				"https://dns.alidns.com/dns-query",
				"tls://dot.pub:853",
				"tls://dns.alidns.com:853",
			},
			Fallback: []string{
				"https://dns.cloudflare.com/dns-query",
				"https://dns.google/dns-query",
				"tls://1.1.1.1:853",
				"tls://8.8.8.8:853",
			},
			FallbackFilter: &clashFallbackFilter{
				GeoIP:     true,
				GeoIPCode: "CN",
				IPCIDR: []string{
					"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10",
					"127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12",
					"192.168.0.0/16", "224.0.0.0/4", "240.0.0.0/4",
				},
				Domain: []string{
					"+.google.com", "+.facebook.com", "+.youtube.com",
					"+.githubusercontent.com", "+.googlevideo.com", "+.googleapis.cn",
				},
			},
			FakeIPFilter: []string{
				"*.lan", "*.local", "*.localhost", "*.test",
				"localhost.ptlogin2.qq.com",
				"+.stun.*.*", "+.stun.*.*.*", "+.stun.*.*.*.*",
				"lens.l.google.com", "*.srv.nintendo.net",
				"+.stun.playstation.net", "xbox.*.*.microsoft.com",
				"*.*.xboxlive.com", "+.msftncsi.com", "+.msftconnecttest.com",
			},
		},
		Proxies: make([]interface{}, 0),
		ProxyGroups: []clashGroup{
			{
				Name:    templateAppNamePlaceholder,
				Type:    "select",
				Proxies: []string{"自动选择", "故障转移", clashAutoProxyPlaceholder, "DIRECT"},
			},
			{
				Name:      "自动选择",
				Type:      "url-test",
				Proxies:   []string{clashAutoProxyPlaceholder},
				URL:       "http://www.gstatic.com/generate_204",
				Interval:  300,
				Tolerance: 50,
			},
			{
				Name:     "故障转移",
				Type:     "fallback",
				Proxies:  []string{clashAutoProxyPlaceholder},
				URL:      "http://www.gstatic.com/generate_204",
				Interval: 300,
			},
		},
		Rules: buildClashRules(),
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "proxies: []\nproxy-groups: []\nrules: []\n"
	}
	return string(data)
}

func defaultSingboxTemplate() string {
	config := map[string]interface{}{
		"dns": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"outbound": []string{"any"},
					"server":   "local",
				},
				{
					"clash_mode": "global",
					"server":     "remote",
				},
				{
					"clash_mode": "direct",
					"server":     "local",
				},
				{
					"rule_set": []string{"geosite-cn"},
					"server":   "local",
				},
			},
			"servers": []map[string]interface{}{
				{
					"address": "https://1.1.1.1/dns-query",
					"detour":  "节点选择",
					"tag":     "remote",
				},
				{
					"address": "https://223.5.5.5/dns-query",
					"detour":  "direct",
					"tag":     "local",
				},
				{
					"address": "rcode://success",
					"tag":     "block",
				},
			},
			"strategy": "prefer_ipv4",
		},
		"experimental": map[string]interface{}{
			"cache_file": map[string]interface{}{
				"enabled":      true,
				"path":         "cache.db",
				"cache_id":     "cache_db",
				"store_fakeip": true,
			},
		},
		"inbounds": []map[string]interface{}{
			{
				"auto_route":                 true,
				"domain_strategy":            "prefer_ipv4",
				"endpoint_independent_nat":   true,
				"address":                    []string{"172.19.0.1/30", "2001:0470:f9da:fdfa::1/64"},
				"mtu":                        9000,
				"sniff":                      true,
				"sniff_override_destination": true,
				"stack":                      "system",
				"strict_route":               true,
				"type":                       "tun",
			},
			{
				"domain_strategy":            "prefer_ipv4",
				"listen":                     "127.0.0.1",
				"listen_port":                2333,
				"sniff":                      true,
				"sniff_override_destination": true,
				"tag":                        "socks-in",
				"type":                       "socks",
				"users":                      []interface{}{},
			},
			{
				"domain_strategy":            "prefer_ipv4",
				"listen":                     "127.0.0.1",
				"listen_port":                2334,
				"sniff":                      true,
				"sniff_override_destination": true,
				"tag":                        "mixed-in",
				"type":                       "mixed",
				"users":                      []interface{}{},
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"tag":     "节点选择",
				"type":    "selector",
				"default": "自动选择",
				"outbounds": []string{
					"自动选择",
					singboxAutoOutboundsPlaceholder,
				},
			},
			{
				"tag":  "direct",
				"type": "direct",
			},
			{
				"tag":  "block",
				"type": "block",
			},
			{
				"tag":  "dns-out",
				"type": "dns",
			},
			{
				"tag":  "自动选择",
				"type": "urltest",
				"outbounds": []string{
					singboxAutoOutboundsPlaceholder,
				},
			},
		},
		"route": map[string]interface{}{
			"auto_detect_interface": true,
			"rules": []map[string]interface{}{
				{
					"outbound": "dns-out",
					"protocol": "dns",
				},
				{
					"clash_mode": "direct",
					"outbound":   "direct",
				},
				{
					"clash_mode": "global",
					"outbound":   "节点选择",
				},
				{
					"ip_is_private": true,
					"outbound":      "direct",
				},
				{
					"rule_set": []string{"geosite-cn", "geoip-cn"},
					"outbound": "direct",
				},
			},
			"rule_set": []map[string]interface{}{
				{
					"tag":             "geosite-cn",
					"type":            "remote",
					"format":          "binary",
					"url":             "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-cn.srs",
					"download_detour": "自动选择",
				},
				{
					"tag":             "geoip-cn",
					"type":            "remote",
					"format":          "binary",
					"url":             "https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set/geoip-cn.srs",
					"download_detour": "自动选择",
				},
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "{\n  \"outbounds\": [],\n  \"route\": {}\n}"
	}
	return string(data)
}

func defaultSurgeTemplate() string {
	return strings.TrimSpace(`
#!MANAGED-CONFIG interval=43200 strict=true
# Surge 的规则配置手册: https://manual.nssurge.com/

[General]
loglevel = notify
doh-server = https://doh.pub/dns-query
dns-server = 223.5.5.5, 114.114.114.114
tun-excluded-routes = 0.0.0.0/8, 10.0.0.0/8, 100.64.0.0/10, 127.0.0.0/8, 169.254.0.0/16, 172.16.0.0/12, 192.0.0.0/24, 192.0.2.0/24, 192.168.0.0/16, 192.88.99.0/24, 198.51.100.0/24, 203.0.113.0/24, 224.0.0.0/4, 255.255.255.255/32
skip-proxy = localhost, *.local, injections.adguard.org, local.adguard.org, captive.apple.com, guzzoni.apple.com, 0.0.0.0/8, 10.0.0.0/8, 17.0.0.0/8, 100.64.0.0/10, 127.0.0.0/8, 169.254.0.0/16, 172.16.0.0/12, 192.0.0.0/24, 192.0.2.0/24, 192.168.0.0/16, 192.88.99.0/24, 198.18.0.0/15, 198.51.100.0/24, 203.0.113.0/24, 224.0.0.0/4, 240.0.0.0/4, 255.255.255.255/32
wifi-assist = true
allow-wifi-access = true
wifi-access-http-port = 6152
wifi-access-socks5-port = 6153
http-listen = 0.0.0.0:6152
socks5-listen = 0.0.0.0:6153
external-controller-access = surgepasswd@0.0.0.0:6170
replica = false
tls-provider = openssl
network-framework = false
exclude-simple-hostnames = true
ipv6 = true
test-timeout = 4
proxy-test-url = http://www.gstatic.com/generate_204
geoip-maxmind-url = https://unpkg.zhimg.com/rulestatic@1.0.1/Country.mmdb

[Replica]
hide-apple-request = true
hide-crashlytics-request = true
use-keyword-filter = false
hide-udp = false

[Proxy]
$proxies

[Proxy Group]
Proxy = select, auto, fallback, $proxy_group
auto = url-test, $proxy_group, url=http://www.gstatic.com/generate_204, interval=43200
fallback = fallback, $proxy_group, url=http://www.gstatic.com/generate_204, interval=43200

[Rule]
$rules

[URL Rewrite]
^https?://(www.)?(g|google).cn https://www.google.com 302
`) + "\n"
}

func defaultSurfboardTemplate() string {
	return strings.TrimSpace(`
#!MANAGED-CONFIG interval=43200 strict=true

[General]
loglevel = notify
ipv6 = false
skip-proxy = localhost, *.local, injections.adguard.org, local.adguard.org, 0.0.0.0/8, 10.0.0.0/8, 17.0.0.0/8, 100.64.0.0/10, 127.0.0.0/8, 169.254.0.0/16, 172.16.0.0/12, 192.0.0.0/24, 192.0.2.0/24, 192.168.0.0/16, 192.88.99.0/24, 198.18.0.0/15, 198.51.100.0/24, 203.0.113.0/24, 224.0.0.0/4, 240.0.0.0/4, 255.255.255.255/32
tls-provider = default
show-error-page-for-reject = true
dns-server = 223.6.6.6, 119.29.29.29, 119.28.28.28
test-timeout = 5
internet-test-url = http://bing.com
proxy-test-url = http://bing.com

[Proxy]
$proxies

[Proxy Group]
Proxy = select, auto, fallback, $proxy_group
auto = url-test, $proxy_group, url=http://www.gstatic.com/generate_204, interval=43200
fallback = fallback, $proxy_group, url=http://www.gstatic.com/generate_204, interval=43200

[Rule]
$rules
`) + "\n"
}
