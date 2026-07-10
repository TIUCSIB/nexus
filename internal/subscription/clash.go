package subscription

import (
	"fmt"
	"strings"

	"nexus/internal/model"

	"gopkg.in/yaml.v3"
)

// clashConfig 是 Clash/Clash.Meta 配置文件的顶层结构。
type clashConfig struct {
	MixedPort         int              `yaml:"mixed-port"`
	AllowLan          bool             `yaml:"allow-lan"`
	BindAddress       string           `yaml:"bind-address,omitempty"`
	Mode              string           `yaml:"mode"`
	LogLevel          string           `yaml:"log-level"`
	ExternalController string          `yaml:"external-controller,omitempty"`
	UnifiedDelay      bool             `yaml:"unified-delay,omitempty"`
	TCPConcurrent     bool             `yaml:"tcp-concurrent,omitempty"`
	DNS               *clashDNS        `yaml:"dns,omitempty"`
	Proxies           []interface{}    `yaml:"proxies"`
	ProxyGroups       []clashGroup     `yaml:"proxy-groups"`
	Rules             []string         `yaml:"rules"`
}

type clashDNS struct {
	Enable           bool              `yaml:"enable"`
	IPv6             bool              `yaml:"ipv6"`
	DefaultNameserver []string         `yaml:"default-nameserver"`
	EnhancedMode     string            `yaml:"enhanced-mode"`
	FakeIPRange      string            `yaml:"fake-ip-range"`
	UseHosts         bool              `yaml:"use-hosts"`
	NameserverPolicy map[string]string `yaml:"nameserver-policy,omitempty"`
	Nameserver       []string          `yaml:"nameserver"`
	Fallback         []string          `yaml:"fallback,omitempty"`
	FallbackFilter   *clashFallbackFilter `yaml:"fallback-filter,omitempty"`
	FakeIPFilter     []string          `yaml:"fake-ip-filter,omitempty"`
}

type clashFallbackFilter struct {
	GeoIP     bool     `yaml:"geoip"`
	GeoIPCode string   `yaml:"geoip-code"`
	IPCIDR    []string `yaml:"ipcidr,omitempty"`
	Domain    []string `yaml:"domain,omitempty"`
}

type clashGroup struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Proxies []string `yaml:"proxies"`
	URL     string   `yaml:"url,omitempty"`
	Interval int     `yaml:"interval,omitempty"`
	Tolerance int    `yaml:"tolerance,omitempty"`
}

// GenerateClash 生成 Clash/Mihomo 格式的 YAML 配置（Xboard 风格）。
func GenerateClash(nodes []model.Node, user model.User, appName string) ([]byte, error) {
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
			Enable:           true,
			IPv6:             false,
			DefaultNameserver: []string{"223.5.5.5", "119.29.29.29"},
			EnhancedMode:     "fake-ip",
			FakeIPRange:      "198.18.0.1/16",
			UseHosts:         true,
			NameserverPolicy: map[string]string{
				"+.google.com":         "https://dns.cloudflare.com/dns-query",
				"+.googleapis.com":     "https://dns.cloudflare.com/dns-query",
				"+.googleapis.cn":      "https://dns.cloudflare.com/dns-query",
				"+.googlevideo.com":    "https://dns.cloudflare.com/dns-query",
				"+.gstatic.com":        "https://dns.cloudflare.com/dns-query",
				"+.youtube.com":        "https://dns.cloudflare.com/dns-query",
				"+.youtu.be":           "https://dns.cloudflare.com/dns-query",
				"+.facebook.com":       "https://dns.cloudflare.com/dns-query",
				"+.twitter.com":        "https://dns.cloudflare.com/dns-query",
				"+.x.com":              "https://dns.cloudflare.com/dns-query",
				"+.github.com":         "https://dns.cloudflare.com/dns-query",
				"+.githubusercontent.com": "https://dns.cloudflare.com/dns-query",
				"+.openai.com":         "https://dns.cloudflare.com/dns-query",
				"+.chatgpt.com":        "https://dns.cloudflare.com/dns-query",
				"+.anthropic.com":      "https://dns.cloudflare.com/dns-query",
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
Proxies:     make([]interface{}, 0),
			ProxyGroups: make([]clashGroup, 0),
			Rules:       buildClashRules(),
		}

	// 构建代理
	nodeNames := make([]string, 0)

	for _, node := range nodes {
		if node.Status != 1 {
			continue
		}
		params := ParseNodeParams(node.ConfigJSON, node.NetworkSettings)
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

	cfg = buildProxyGroups(cfg, nodeNames, appName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	// 替换 $app_name 占位符（与 Xboard 一致）
	groupName := appName
	if groupName == "" {
		groupName = "$app_name"
	}
	yamlStr := strings.ReplaceAll(string(data), "$app_name", groupName)
	return []byte(yamlStr), nil
}

// buildClashRules 生成完整的 Clash 规则列表，使用 $app_name 占位符
// YAML 输出后在 GenerateClash 中替换为实际的站点名称
func buildClashRules() []string {
	rules := []string{
		// 广告拦截
		"DOMAIN-KEYWORD,admarvel,REJECT",
		"DOMAIN-KEYWORD,admaster,REJECT",
		"DOMAIN-KEYWORD,adsage,REJECT",
		"DOMAIN-KEYWORD,adsmogo,REJECT",
		"DOMAIN-KEYWORD,adsrvmedia,REJECT",
		"DOMAIN-KEYWORD,adwords,REJECT",
		"DOMAIN-KEYWORD,adservice,REJECT",
		"DOMAIN-KEYWORD,domob,REJECT",
		"DOMAIN-KEYWORD,duomeng,REJECT",
		"DOMAIN-KEYWORD,dwtrack,REJECT",
		"DOMAIN-KEYWORD,guanggao,REJECT",
		"DOMAIN-KEYWORD,lianmeng,REJECT",
		"DOMAIN-KEYWORD,omgmta,REJECT",
		"DOMAIN-KEYWORD,openx,REJECT",
		"DOMAIN-KEYWORD,partnerad,REJECT",
		"DOMAIN-KEYWORD,supersonicads,REJECT",
		"DOMAIN-KEYWORD,umeng,REJECT",
		"DOMAIN-KEYWORD,zjtoolbar,REJECT",
		"DOMAIN-SUFFIX,appsflyer.com,REJECT",
		"DOMAIN-SUFFIX,doubleclick.net,REJECT",
		"DOMAIN-SUFFIX,mmstat.com,REJECT",

		// 本地直连
		"DOMAIN-SUFFIX,local,DIRECT",
		"DOMAIN-SUFFIX,localhost,DIRECT",
		"IP-CIDR,10.0.0.0/8,DIRECT,no-resolve",
		"IP-CIDR,17.0.0.0/8,DIRECT,no-resolve",
		"IP-CIDR,100.64.0.0/10,DIRECT,no-resolve",
		"IP-CIDR,127.0.0.0/8,DIRECT,no-resolve",
		"IP-CIDR,172.16.0.0/12,DIRECT,no-resolve",
		"IP-CIDR,192.168.0.0/16,DIRECT,no-resolve",
		"IP-CIDR,198.18.0.0/16,DIRECT,no-resolve",
		"IP-CIDR,224.0.0.0/4,DIRECT,no-resolve",
		"IP-CIDR6,::1/128,DIRECT,no-resolve",
		"IP-CIDR6,fc00::/7,DIRECT,no-resolve",
		"IP-CIDR6,fe80::/10,DIRECT,no-resolve",

		// Apple 服务
		"DOMAIN-SUFFIX,apps.apple.com,Proxy",
		"DOMAIN-SUFFIX,itunes.apple.com,Proxy",
		"DOMAIN-SUFFIX,blobstore.apple.com,Proxy",
		"DOMAIN,safebrowsing.urlsec.qq.com,DIRECT",
		"DOMAIN-SUFFIX,apple.com,DIRECT",
		"DOMAIN-SUFFIX,apple-cloudkit.com,DIRECT",
		"DOMAIN-SUFFIX,icloud.com,DIRECT",
		"DOMAIN-SUFFIX,icloud-content.com,DIRECT",
		"DOMAIN-SUFFIX,mzstatic.com,DIRECT",
		"DOMAIN-SUFFIX,aaplimg.com,DIRECT",
		"DOMAIN-SUFFIX,cdn-apple.com,DIRECT",
		"DOMAIN-SUFFIX,akadns.net,DIRECT",

		// 国内网站直连
		"DOMAIN-KEYWORD,baidu,DIRECT",
		"DOMAIN-KEYWORD,alibaba,DIRECT",
		"DOMAIN-KEYWORD,alicdn,DIRECT",
		"DOMAIN-KEYWORD,alipay,DIRECT",
		"DOMAIN-KEYWORD,taobao,DIRECT",
		"DOMAIN-KEYWORD,tencent,DIRECT",
		"DOMAIN-KEYWORD,bilibili,DIRECT",
		"DOMAIN-KEYWORD,weibo,DIRECT",
		"DOMAIN-KEYWORD,douyin,DIRECT",
		"DOMAIN-KEYWORD,bytedance,DIRECT",
		"DOMAIN-KEYWORD,xiaomi,DIRECT",
		"DOMAIN-KEYWORD,huawei,DIRECT",
		"DOMAIN-KEYWORD,netease,DIRECT",
		"DOMAIN-KEYWORD,meituan,DIRECT",
		"DOMAIN-KEYWORD,pinduoduo,DIRECT",
		"DOMAIN-KEYWORD,kuaishou,DIRECT",
		"DOMAIN-KEYWORD,jingdong,DIRECT",
		"DOMAIN-KEYWORD,officecdn,DIRECT",
		"DOMAIN-SUFFIX,qq.com,DIRECT",
		"DOMAIN-SUFFIX,weixin.com,DIRECT",
		"DOMAIN-SUFFIX,wechat.com,DIRECT",
		"DOMAIN-SUFFIX,gtimg.com,DIRECT",
		"DOMAIN-SUFFIX,qcloud.com,DIRECT",
		"DOMAIN-SUFFIX,myqcloud.com,DIRECT",
		"DOMAIN-SUFFIX,qpic.cn,DIRECT",
		"DOMAIN-SUFFIX,tenpay.com,DIRECT",
		"DOMAIN-SUFFIX,tmall.com,DIRECT",
		"DOMAIN-SUFFIX,jd.com,DIRECT",
		"DOMAIN-SUFFIX,360buyimg.com,DIRECT",
		"DOMAIN-SUFFIX,iqiyi.com,DIRECT",
		"DOMAIN-SUFFIX,youku.com,DIRECT",
		"DOMAIN-SUFFIX,ykimg.com,DIRECT",
		"DOMAIN-SUFFIX,tudou.com,DIRECT",
		"DOMAIN-SUFFIX,acfun.tv,DIRECT",
		"DOMAIN-SUFFIX,hdslb.com,DIRECT",
		"DOMAIN-SUFFIX,sohu.com,DIRECT",
		"DOMAIN-SUFFIX,sogou.com,DIRECT",
		"DOMAIN-SUFFIX,zhihu.com,DIRECT",
		"DOMAIN-SUFFIX,zhimg.com,DIRECT",
		"DOMAIN-SUFFIX,douban.com,DIRECT",
		"DOMAIN-SUFFIX,doubanio.com,DIRECT",
		"DOMAIN-SUFFIX,163.com,DIRECT",
		"DOMAIN-SUFFIX,126.com,DIRECT",
		"DOMAIN-SUFFIX,126.net,DIRECT",
		"DOMAIN-SUFFIX,127.net,DIRECT",
		"DOMAIN-SUFFIX,yeah.net,DIRECT",
		"DOMAIN-SUFFIX,sina.com,DIRECT",
		"DOMAIN-SUFFIX,sinaimg.cn,DIRECT",
		"DOMAIN-SUFFIX,ximalaya.com,DIRECT",
		"DOMAIN-SUFFIX,xmcdn.com,DIRECT",
		"DOMAIN-SUFFIX,csdn.net,DIRECT",
		"DOMAIN-SUFFIX,gitee.com,DIRECT",
		"DOMAIN-SUFFIX,jianshu.com,DIRECT",
		"DOMAIN-SUFFIX,cnblogs.com,DIRECT",
		"DOMAIN-SUFFIX,oschina.net,DIRECT",
		"DOMAIN-SUFFIX,ele.me,DIRECT",
		"DOMAIN-SUFFIX,ctrip.com,DIRECT",
		"DOMAIN-SUFFIX,suning.com,DIRECT",
		"DOMAIN-SUFFIX,dianping.com,DIRECT",
		"DOMAIN-SUFFIX,amap.com,DIRECT",
		"DOMAIN-SUFFIX,autonavi.com,DIRECT",
		"DOMAIN-SUFFIX,mi.com,DIRECT",
		"DOMAIN-SUFFIX,miui.com,DIRECT",
		"DOMAIN-SUFFIX,ifeng.com,DIRECT",
		"DOMAIN-SUFFIX,youdao.com,DIRECT",
		"DOMAIN-SUFFIX,iciba.com,DIRECT",
		"DOMAIN-SUFFIX,xunlei.com,DIRECT",
		"DOMAIN-SUFFIX,smzdm.com,DIRECT",
		"DOMAIN-SUFFIX,sspai.com,DIRECT",
		"DOMAIN-SUFFIX,36kr.com,DIRECT",
		"DOMAIN-SUFFIX,speedtest.net,DIRECT",
		"DOMAIN-SUFFIX,microsoft.com,DIRECT",
		"DOMAIN-SUFFIX,microsoftonline.com,DIRECT",
		"DOMAIN-SUFFIX,office.com,DIRECT",
		"DOMAIN-SUFFIX,office365.com,DIRECT",
		"DOMAIN-SUFFIX,windows.com,DIRECT",
		"DOMAIN-SUFFIX,windowsupdate.com,DIRECT",
		"DOMAIN-SUFFIX,live.com,DIRECT",
		"DOMAIN-SUFFIX,msn.com,DIRECT",
		"DOMAIN-SUFFIX,cn,DIRECT",
		"DOMAIN-KEYWORD,-cn,DIRECT",

		// 国际网站走代理
		"DOMAIN-KEYWORD,google,Proxy",
		"DOMAIN-KEYWORD,gmail,Proxy",
		"DOMAIN-KEYWORD,youtube,Proxy",
		"DOMAIN-KEYWORD,facebook,Proxy",
		"DOMAIN-KEYWORD,twitter,Proxy",
		"DOMAIN-KEYWORD,instagram,Proxy",
		"DOMAIN-KEYWORD,whatsapp,Proxy",
		"DOMAIN-KEYWORD,telegram,Proxy",
		"DOMAIN-KEYWORD,github,Proxy",
		"DOMAIN-KEYWORD,blogspot,Proxy",
		"DOMAIN-KEYWORD,dropbox,Proxy",
		"DOMAIN-KEYWORD,wikipedia,Proxy",
		"DOMAIN-KEYWORD,pinterest,Proxy",
		"DOMAIN-KEYWORD,discord,Proxy",
		"DOMAIN-KEYWORD,openai,Proxy",
		"DOMAIN-KEYWORD,anthropic,Proxy",
		"DOMAIN-KEYWORD,netflix,Proxy",
		"DOMAIN-KEYWORD,spotify,Proxy",
		"DOMAIN-KEYWORD,amazon,Proxy",
		"DOMAIN-SUFFIX,t.co,Proxy",
		"DOMAIN-SUFFIX,x.com,Proxy",
		"DOMAIN-SUFFIX,twimg.com,Proxy",
		"DOMAIN-SUFFIX,fb.me,Proxy",
		"DOMAIN-SUFFIX,fbcdn.net,Proxy",
		"DOMAIN-SUFFIX,youtu.be,Proxy",
		"DOMAIN-SUFFIX,ytimg.com,Proxy",
		"DOMAIN-SUFFIX,gstatic.com,Proxy",
		"DOMAIN-SUFFIX,ggpht.com,Proxy",
		"DOMAIN-SUFFIX,googlevideo.com,Proxy",
		"DOMAIN-SUFFIX,v2ex.com,Proxy",
		"DOMAIN-SUFFIX,medium.com,Proxy",
		"DOMAIN-SUFFIX,reddit.com,Proxy",
		"DOMAIN-SUFFIX,redd.it,Proxy",
		"DOMAIN-SUFFIX,imgur.com,Proxy",
		"DOMAIN-SUFFIX,pixiv.net,Proxy",
		"DOMAIN-SUFFIX,nytimes.com,Proxy",
		"DOMAIN-SUFFIX,nyt.com,Proxy",
		"DOMAIN-SUFFIX,bbc.com,Proxy",
		"DOMAIN-SUFFIX,bbc.co.uk,Proxy",
		"DOMAIN-SUFFIX,steamcommunity.com,Proxy",
		"DOMAIN-SUFFIX,twitch.tv,Proxy",
		"DOMAIN-SUFFIX,vimeo.com,Proxy",
		"DOMAIN-SUFFIX,tumblr.com,Proxy",
		"DOMAIN-SUFFIX,linkedin.com,Proxy",
		"DOMAIN-SUFFIX,licdn.com,Proxy",
		"DOMAIN-SUFFIX,mega.nz,Proxy",
		"DOMAIN-SUFFIX,archive.org,Proxy",
		"DOMAIN-SUFFIX,wikimedia.org,Proxy",
		"DOMAIN-SUFFIX,soundcloud.com,Proxy",

		// Telegram IP 段
		"IP-CIDR,91.108.4.0/22,Proxy,no-resolve",
		"IP-CIDR,91.108.8.0/21,Proxy,no-resolve",
		"IP-CIDR,91.108.12.0/22,Proxy,no-resolve",
		"IP-CIDR,91.108.16.0/22,Proxy,no-resolve",
		"IP-CIDR,91.108.56.0/22,Proxy,no-resolve",
		"IP-CIDR,149.154.160.0/20,Proxy,no-resolve",
		"IP-CIDR6,2001:67c:4e8::/48,Proxy,no-resolve",
		"IP-CIDR6,2001:b28:f23d::/48,Proxy,no-resolve",
		"IP-CIDR6,2001:b28:f23f::/48,Proxy,no-resolve",

		// 结尾规则
		"GEOIP,CN,DIRECT",
		"MATCH,Proxy",
	}

	// 将规则中的 Proxy 替换为 $app_name 占位符
	for i, r := range rules {
		rules[i] = strings.ReplaceAll(r, ",Proxy,", ",$app_name,")
		rules[i] = strings.ReplaceAll(rules[i], ",Proxy", ",$app_name")
	}
	return rules
}

// buildProxyGroups 构建三个代理分组
func buildProxyGroups(cfg clashConfig, allNames []string, appName string) clashConfig {
	if len(allNames) == 0 {
		return cfg
	}

	// 主选择分组名称（使用 $app_name 占位符，YAML 输出后替换为实际名称）
	groupName := "$app_name"

	cfg.ProxyGroups = []clashGroup{
		{
			Name:    groupName,
			Type:    "select",
			Proxies: append([]string{"自动选择", "故障转移", "DIRECT"}, allNames...),
		},
		{
			Name:     "自动选择",
			Type:     "url-test",
			Proxies:  allNames,
			URL:      "http://www.gstatic.com/generate_204",
			Interval: 300,
			Tolerance: 50,
		},
		{
			Name:     "故障转移",
			Type:     "fallback",
			Proxies:  allNames,
			URL:      "http://www.gstatic.com/generate_204",
			Interval: 300,
		},
	}

	return cfg
}

// buildClashVLESS 构建 VLESS 代理配置（Xboard 风格）
func buildClashVLESS(node model.Node, user model.User, p NodeParams) map[string]interface{} {
	proxy := map[string]interface{}{
		"name":       node.Name,
		"type":       "vless",
		"server":     node.Address,
		"port":       node.Port,
		"uuid":       user.UUID,
		"alterId":    0,
		"cipher":     "auto",
		"udp":        true,
		"encryption": "none",
		"tls":        true,
	}

	// Flow — only add when explicitly set (Xboard behavior)
	flow := node.FlowControl
	if flow != "" && flow != "none" {
		proxy["flow"] = flow
	} else {
		proxy["flow"] = nil
	}

	// Transport
	if node.Transport != "" && node.Transport != "tcp" {
		proxy["network"] = node.Transport
	}

	// TLS servername（Reality 时为 handshake dest server）
	serverName := p.ServerName
	if serverName == "" {
		serverName = p.HandshakeHost
	}
	if serverName != "" {
		proxy["servername"] = serverName
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

// buildClashHysteria2 构建 Hysteria2 代理配置
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

// buildClashTUIC 构建 TUIC 代理配置
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
		"name":               node.Name,
		"type":               "tuic",
		"server":             node.Address,
		"port":               node.Port,
		"uuid":               user.UUID,
		"password":           password,
		"congestion-control": congestion,
		"tls":                true,
		"udp":                true,
	}

	if p.ServerName != "" {
		proxy["sni"] = p.ServerName
	}

	return proxy
}