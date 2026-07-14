package subscription

import (
	"fmt"
	"strings"

	"nexus/internal/model"

	"gopkg.in/yaml.v3"
)

// clashConfig 是 Clash/Clash.Meta 配置文件的顶层结构。
type clashConfig struct {
	MixedPort          int           `yaml:"mixed-port"`
	AllowLan           bool          `yaml:"allow-lan"`
	BindAddress        string        `yaml:"bind-address,omitempty"`
	Mode               string        `yaml:"mode"`
	LogLevel           string        `yaml:"log-level"`
	ExternalController string        `yaml:"external-controller,omitempty"`
	UnifiedDelay       bool          `yaml:"unified-delay,omitempty"`
	TCPConcurrent      bool          `yaml:"tcp-concurrent,omitempty"`
	DNS                *clashDNS     `yaml:"dns,omitempty"`
	Proxies            []interface{} `yaml:"proxies"`
	ProxyGroups        []clashGroup  `yaml:"proxy-groups"`
	Rules              []string      `yaml:"rules"`
}

type clashDNS struct {
	Enable            bool                 `yaml:"enable"`
	IPv6              bool                 `yaml:"ipv6"`
	DefaultNameserver []string             `yaml:"default-nameserver"`
	EnhancedMode      string               `yaml:"enhanced-mode"`
	FakeIPRange       string               `yaml:"fake-ip-range"`
	UseHosts          bool                 `yaml:"use-hosts"`
	NameserverPolicy  map[string]string    `yaml:"nameserver-policy,omitempty"`
	Nameserver        []string             `yaml:"nameserver"`
	Fallback          []string             `yaml:"fallback,omitempty"`
	FallbackFilter    *clashFallbackFilter `yaml:"fallback-filter,omitempty"`
	FakeIPFilter      []string             `yaml:"fake-ip-filter,omitempty"`
}

type clashFallbackFilter struct {
	GeoIP     bool     `yaml:"geoip"`
	GeoIPCode string   `yaml:"geoip-code"`
	IPCIDR    []string `yaml:"ipcidr,omitempty"`
	Domain    []string `yaml:"domain,omitempty"`
}

type clashGroup struct {
	Name      string   `yaml:"name"`
	Type      string   `yaml:"type"`
	Proxies   []string `yaml:"proxies"`
	URL       string   `yaml:"url,omitempty"`
	Interval  int      `yaml:"interval,omitempty"`
	Tolerance int      `yaml:"tolerance,omitempty"`
}

// GenerateClash 生成 Clash/Mihomo 格式的 YAML 配置（Xboard 风格）。
func GenerateClash(nodes []model.Node, user model.User, appName string) ([]byte, error) {
	return generateClashByTemplate(nodes, user, appName, SettingSubscribeTemplateClash)
}

// GenerateClashMeta 生成 Clash Meta 独立模板格式。
func GenerateClashMeta(nodes []model.Node, user model.User, appName string) ([]byte, error) {
	return generateClashByTemplate(nodes, user, appName, SettingSubscribeTemplateClashMeta)
}

// GenerateStash 生成 Stash 独立模板格式。
func GenerateStash(nodes []model.Node, user model.User, appName string) ([]byte, error) {
	return generateClashByTemplate(nodes, user, appName, SettingSubscribeTemplateStash)
}

func generateClashByTemplate(nodes []model.Node, user model.User, appName string, templateKey string) ([]byte, error) {
	templateContent := GetSubscriptionTemplate(templateKey)
	var cfg map[string]interface{}
	if err := yaml.Unmarshal([]byte(templateContent), &cfg); err != nil {
		return nil, err
	}

	proxies := make([]interface{}, 0)
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
			proxies = append(proxies, proxy)
			nodeNames = append(nodeNames, node.Name)
		}
	}

	cfg["proxies"] = proxies
	if err := applyClashProxyGroups(cfg, nodeNames); err != nil {
		return nil, err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	groupName := appName
	if groupName == "" {
		groupName = templateAppNamePlaceholder
	}
	output := strings.ReplaceAll(string(data), templateAppNamePlaceholder, groupName)
	return []byte(output), nil
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
	groupName := templateAppNamePlaceholder

	cfg.ProxyGroups = []clashGroup{
		{
			Name:    groupName,
			Type:    "select",
			Proxies: append([]string{"自动选择", "故障转移", "DIRECT"}, allNames...),
		},
		{
			Name:      "自动选择",
			Type:      "url-test",
			Proxies:   allNames,
			URL:       "http://www.gstatic.com/generate_204",
			Interval:  300,
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

func applyClashProxyGroups(cfg map[string]interface{}, nodeNames []string) error {
	groupsValue, ok := cfg["proxy-groups"]
	if !ok {
		return fmt.Errorf("Clash 模板缺少 proxy-groups 字段")
	}

	groups, ok := groupsValue.([]interface{})
	if !ok {
		return fmt.Errorf("Clash 模板的 proxy-groups 格式无效")
	}

	for _, item := range groups {
		group, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		proxiesValue, ok := group["proxies"]
		if !ok {
			continue
		}
		proxies, ok := proxiesValue.([]interface{})
		if !ok {
			continue
		}
		group["proxies"] = replaceClashGroupPlaceholders(proxies, nodeNames)
	}

	return nil
}

func replaceClashGroupPlaceholders(proxies []interface{}, nodeNames []string) []interface{} {
	result := make([]interface{}, 0, len(proxies)+len(nodeNames))
	for _, item := range proxies {
		name, ok := item.(string)
		if !ok {
			result = append(result, item)
			continue
		}
		if name == clashAutoProxyPlaceholder {
			for _, nodeName := range nodeNames {
				result = append(result, nodeName)
			}
			continue
		}
		result = append(result, item)
	}
	return result
}

// buildClashVLESS 构建 VLESS 代理配置（Xboard 风格）
func buildClashVLESS(node model.Node, user model.User, p NodeParams) map[string]interface{} {
	proxy := map[string]interface{}{
		"name":       node.Name,
		"type":       "vless",
		"server":     node.Address,
		"port":       node.Port,
		"uuid":       user.UUID,
		"udp":        true,
		"encryption": "none",
	}

	// 传输层：tcp 可不写 network；其它传输显式写出
	network := strings.ToLower(strings.TrimSpace(node.Transport))
	if network == "" {
		network = "tcp"
	}
	if network != "tcp" {
		proxy["network"] = network
	}

	// Flow — 仅当显式设置且不为 none 时添加
	flow := node.FlowControl
	if flow != "" && flow != "none" {
		proxy["flow"] = flow
	}

	// TLS / Reality — 严格跟随节点 security，none 时明确关闭 TLS
	security := strings.ToLower(strings.TrimSpace(node.Security))
	if security == "" {
		security = "none"
	}
	hasTLS := security == "tls" || security == "reality"
	proxy["tls"] = hasTLS
	if !hasTLS {
		// 无 TLS 时不要带 servername/指纹，避免客户端 UI 误显示 TLS
		return proxy
	}

	// TLS servername（Reality 时为 handshake dest server）
	serverName := p.ServerName
	if serverName == "" {
		serverName = p.HandshakeHost
	}
	if serverName != "" {
		proxy["servername"] = serverName
	}
	if p.AllowInsecure {
		proxy["skip-cert-verify"] = true
	}

	// Reality 配置
	if security == "reality" && p.PublicKey != "" {
		realityOpts := map[string]interface{}{
			"public-key": p.PublicKey,
		}
		if p.ShortID != "" {
			realityOpts["short-id"] = p.ShortID
		}
		if p.HandshakeHost != "" {
			realityOpts["servername"] = p.HandshakeHost
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
		"udp":      true,
	}

	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}
	if sni != "" {
		proxy["sni"] = sni
	}
	if p.AllowInsecure {
		proxy["skip-cert-verify"] = true
	}

	if p.UpMbps > 0 {
		proxy["up"] = p.UpMbps
	}
	if p.DownMbps > 0 {
		proxy["down"] = p.DownMbps
	}

	if p.ObfsEnabled && p.ObfsPassword != "" {
		obfsType := p.ObfsType
		if obfsType == "" {
			obfsType = "salamander"
		}
		proxy["obfs"] = obfsType
		proxy["obfs-password"] = p.ObfsPassword
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

	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}
	if sni != "" {
		proxy["sni"] = sni
	}
	if p.AllowInsecure {
		proxy["skip-cert-verify"] = true
	}

	return proxy
}
