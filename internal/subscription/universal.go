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

	for _, node := range nodes {
		if node.Status != 1 {
			continue
		}
		params := ParseNodeParams(node.ConfigJSON, node.NetworkSettings)
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
			lines = append(lines, uri)
		}
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
	return generateTextSubscription(nodes, user, SettingSubscribeTemplateSurge)
}

// GenerateSurfboard 生成 Surfboard 配置格式。
func GenerateSurfboard(nodes []model.Node, user model.User) ([]byte, error) {
	return generateTextSubscription(nodes, user, SettingSubscribeTemplateSurfboard)
}

func generateTextSubscription(nodes []model.Node, user model.User, settingKey string) ([]byte, error) {
	proxyLines, proxyNames := buildTextSubscriptionProxies(nodes, user)
	templateContent := GetSubscriptionTemplate(settingKey)
	content := strings.ReplaceAll(templateContent, templateProxiesPlaceholder, strings.Join(proxyLines, "\n"))
	content = strings.ReplaceAll(content, templateProxyGroupPlaceholder, buildTextSubscriptionProxyGroup(proxyNames))
	content = strings.ReplaceAll(content, templateRulesPlaceholder, buildTextSubscriptionRules())
	content = strings.ReplaceAll(content, templateAppNamePlaceholder, GetSubscriptionAppName())
	return []byte(content), nil
}

func buildTextSubscriptionProxies(nodes []model.Node, user model.User) ([]string, []string) {
	proxyLines := make([]string, 0)
	proxyNames := make([]string, 0)

	for _, node := range nodes {
		if node.Status != 1 {
			continue
		}
		params := ParseNodeParams(node.ConfigJSON, node.NetworkSettings)
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
			proxyLines = append(proxyLines, line)
			proxyNames = append(proxyNames, node.Name)
		}
	}

	return proxyLines, proxyNames
}

func buildTextSubscriptionProxyGroup(proxyNames []string) string {
	if len(proxyNames) == 0 {
		return ""
	}
	return fmt.Sprintf("Proxy = select, %s", strings.Join(proxyNames, ", "))
}

func buildTextSubscriptionRules() string {
	return strings.Join([]string{
		"# 自定义规则",
		"## 您可以在此处插入自定义规则",
		"",
		"# Google 中国服务",
		"DOMAIN-SUFFIX,services.googleapis.cn,Proxy",
		"DOMAIN-SUFFIX,xn--ngstr-lra8j.com,Proxy",
		"",
		"# Apple",
		"DOMAIN,developer.apple.com,Proxy",
		"DOMAIN-SUFFIX,digicert.com,Proxy",
		"USER-AGENT,com.apple.trustd*,Proxy",
		"DOMAIN-SUFFIX,apple-dns.net,Proxy",
		"DOMAIN,testflight.apple.com,Proxy",
		"DOMAIN,sandbox.itunes.apple.com,Proxy",
		"DOMAIN,itunes.apple.com,Proxy",
		"DOMAIN-SUFFIX,apps.apple.com,Proxy",
		"DOMAIN-SUFFIX,blobstore.apple.com,Proxy",
		"DOMAIN,cvws.icloud-content.com,Proxy",
		"DOMAIN,safebrowsing.urlsec.qq.com,DIRECT",
		"DOMAIN,safebrowsing.googleapis.com,DIRECT",
		"USER-AGENT,com.apple.appstored*,DIRECT",
		"USER-AGENT,AppStore*,DIRECT",
		"DOMAIN-SUFFIX,mzstatic.com,DIRECT",
		"DOMAIN-SUFFIX,itunes.apple.com,DIRECT",
		"DOMAIN-SUFFIX,icloud.com,DIRECT",
		"DOMAIN-SUFFIX,icloud-content.com,DIRECT",
		"USER-AGENT,cloudd*,DIRECT",
		"USER-AGENT,*com.apple.WebKit*,DIRECT",
		"USER-AGENT,*com.apple.*,DIRECT",
		"DOMAIN-SUFFIX,me.com,DIRECT",
		"DOMAIN-SUFFIX,aaplimg.com,DIRECT",
		"DOMAIN-SUFFIX,cdn-apple.com,DIRECT",
		"DOMAIN-SUFFIX,akadns.net,DIRECT",
		"DOMAIN-SUFFIX,apple.com,DIRECT",
		"DOMAIN-SUFFIX,apple-cloudkit.com,DIRECT",
		"",
		"# 国内网站",
		"USER-AGENT,MicroMessenger Client*,DIRECT",
		"USER-AGENT,WeChat*,DIRECT",
		"DOMAIN-SUFFIX,126.com,DIRECT",
		"DOMAIN-SUFFIX,126.net,DIRECT",
		"DOMAIN-SUFFIX,127.net,DIRECT",
		"DOMAIN-SUFFIX,163.com,DIRECT",
		"DOMAIN-SUFFIX,360buyimg.com,DIRECT",
		"DOMAIN-SUFFIX,36kr.com,DIRECT",
		"DOMAIN-SUFFIX,acfun.tv,DIRECT",
		"DOMAIN-KEYWORD,alicdn,DIRECT",
		"DOMAIN-KEYWORD,alipay,DIRECT",
		"DOMAIN-KEYWORD,aliyun,DIRECT",
		"DOMAIN-KEYWORD,taobao,DIRECT",
		"DOMAIN-SUFFIX,amap.com,DIRECT",
		"DOMAIN-SUFFIX,autonavi.com,DIRECT",
		"DOMAIN-KEYWORD,baidu,DIRECT",
		"DOMAIN-SUFFIX,bilibili.com,DIRECT",
		"DOMAIN-SUFFIX,csdn.net,DIRECT",
		"DOMAIN-SUFFIX,dianping.com,DIRECT",
		"DOMAIN-SUFFIX,douban.com,DIRECT",
		"DOMAIN-SUFFIX,doubanio.com,DIRECT",
		"DOMAIN-SUFFIX,ele.me,DIRECT",
		"DOMAIN-SUFFIX,gtimg.com,DIRECT",
		"DOMAIN-SUFFIX,iciba.com,DIRECT",
		"DOMAIN-SUFFIX,ifeng.com,DIRECT",
		"DOMAIN-SUFFIX,iqiyi.com,DIRECT",
		"DOMAIN-SUFFIX,jd.com,DIRECT",
		"DOMAIN-SUFFIX,jianshu.com,DIRECT",
		"DOMAIN-SUFFIX,meituan.com,DIRECT",
		"DOMAIN-SUFFIX,microsoft.com,DIRECT",
		"DOMAIN-SUFFIX,microsoftonline.com,DIRECT",
		"DOMAIN-SUFFIX,mi.com,DIRECT",
		"DOMAIN-SUFFIX,miui.com,DIRECT",
		"DOMAIN-SUFFIX,netease.com,DIRECT",
		"DOMAIN-SUFFIX,office.com,DIRECT",
		"DOMAIN-KEYWORD,officecdn,DIRECT",
		"DOMAIN-SUFFIX,office365.com,DIRECT",
		"DOMAIN-SUFFIX,oschina.net,DIRECT",
		"DOMAIN-SUFFIX,qcloud.com,DIRECT",
		"DOMAIN-SUFFIX,qq.com,DIRECT",
		"DOMAIN-SUFFIX,sina.com,DIRECT",
		"DOMAIN-SUFFIX,smzdm.com,DIRECT",
		"DOMAIN-SUFFIX,sogou.com,DIRECT",
		"DOMAIN-SUFFIX,sohu.com,DIRECT",
		"DOMAIN-SUFFIX,speedtest.net,DIRECT",
		"DOMAIN-SUFFIX,sspai.com,DIRECT",
		"DOMAIN-SUFFIX,suning.com,DIRECT",
		"DOMAIN-SUFFIX,taobao.com,DIRECT",
		"DOMAIN-SUFFIX,tencent.com,DIRECT",
		"DOMAIN-SUFFIX,tenpay.com,DIRECT",
		"DOMAIN-SUFFIX,tudou.com,DIRECT",
		"DOMAIN-SUFFIX,weibo.com,DIRECT",
		"DOMAIN-SUFFIX,ximalaya.com,DIRECT",
		"DOMAIN-SUFFIX,xmcdn.com,DIRECT",
		"DOMAIN-SUFFIX,xunlei.com,DIRECT",
		"DOMAIN-SUFFIX,youdao.com,DIRECT",
		"DOMAIN-SUFFIX,youku.com,DIRECT",
		"DOMAIN-SUFFIX,zhihu.com,DIRECT",
		"DOMAIN-SUFFIX,zhimg.com,DIRECT",
		"",
		"# 常见广告",
		"DOMAIN-KEYWORD,admarvel,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,admaster,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,adsage,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,adsmogo,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,adsrvmedia,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,adwords,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,adservice,REJECT-TINYGIF",
		"DOMAIN-SUFFIX,appsflyer.com,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,domob,REJECT-TINYGIF",
		"DOMAIN-SUFFIX,doubleclick.net,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,duomeng,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,dwtrack,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,guanggao,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,lianmeng,REJECT-TINYGIF",
		"DOMAIN-SUFFIX,mmstat.com,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,omgmta,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,openx,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,partnerad,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,supersonicads,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,umeng,REJECT-TINYGIF",
		"DOMAIN-KEYWORD,zjtoolbar,REJECT-TINYGIF",
		"",
		"# 抗 DNS 污染",
		"DOMAIN-KEYWORD,amazon,Proxy",
		"DOMAIN-KEYWORD,google,Proxy",
		"DOMAIN-KEYWORD,gmail,Proxy",
		"DOMAIN-KEYWORD,youtube,Proxy",
		"DOMAIN-KEYWORD,facebook,Proxy",
		"DOMAIN-SUFFIX,fb.me,Proxy",
		"DOMAIN-SUFFIX,fbcdn.net,Proxy",
		"DOMAIN-KEYWORD,twitter,Proxy",
		"DOMAIN-KEYWORD,instagram,Proxy",
		"DOMAIN-KEYWORD,dropbox,Proxy",
		"DOMAIN-SUFFIX,twimg.com,Proxy",
		"DOMAIN-KEYWORD,blogspot,Proxy",
		"DOMAIN-SUFFIX,youtu.be,Proxy",
		"DOMAIN-KEYWORD,github,Proxy",
		"DOMAIN-SUFFIX,github.com,Proxy",
		"DOMAIN-SUFFIX,githubusercontent.com,Proxy",
		"DOMAIN-KEYWORD,openai,Proxy",
		"DOMAIN-KEYWORD,anthropic,Proxy",
		"DOMAIN-SUFFIX,chatgpt.com,Proxy",
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
		"DOMAIN-SUFFIX,mega.nz,Proxy",
		"DOMAIN-SUFFIX,archive.org,Proxy",
		"DOMAIN-SUFFIX,wikimedia.org,Proxy",
		"DOMAIN-SUFFIX,soundcloud.com,Proxy",
		"",
		"# Telegram",
		"DOMAIN-SUFFIX,telegra.ph,Proxy",
		"DOMAIN-SUFFIX,telegram.org,Proxy",
		"IP-CIDR,91.108.4.0/22,Proxy,no-resolve",
		"IP-CIDR,91.108.8.0/21,Proxy,no-resolve",
		"IP-CIDR,91.108.16.0/22,Proxy,no-resolve",
		"IP-CIDR,91.108.56.0/22,Proxy,no-resolve",
		"IP-CIDR,149.154.160.0/20,Proxy,no-resolve",
		"IP-CIDR6,2001:67c:4e8::/48,Proxy,no-resolve",
		"IP-CIDR6,2001:b28:f23d::/48,Proxy,no-resolve",
		"IP-CIDR6,2001:b28:f23f::/48,Proxy,no-resolve",
		"",
		"# LAN",
		"DOMAIN-SUFFIX,local,DIRECT",
		"IP-CIDR,127.0.0.0/8,DIRECT",
		"IP-CIDR,172.16.0.0/12,DIRECT",
		"IP-CIDR,192.168.0.0/16,DIRECT",
		"IP-CIDR,10.0.0.0/8,DIRECT",
		"IP-CIDR,17.0.0.0/8,DIRECT",
		"IP-CIDR,100.64.0.0/10,DIRECT",
		"IP-CIDR,224.0.0.0/4,DIRECT",
		"IP-CIDR6,fe80::/10,DIRECT",
		"",
		"# 剩余未匹配的国内网站",
		"DOMAIN-SUFFIX,cn,DIRECT",
		"DOMAIN-KEYWORD,-cn,DIRECT",
		"",
		"# 最终规则",
		"GEOIP,CN,DIRECT",
		"FINAL,Proxy",
	}, "\n")
}

// ==================== URI 构建 ====================

func buildVlessURI(node model.Node, user model.User, p NodeParams) string {
	// 流控
	flow := node.FlowControl
	if flow == "" {
		flow = "none"
	}

	// 安全性
	security := node.Security
	if security == "" {
		security = "none"
	}

	q := url.Values{}
	if flow != "" && flow != "none" {
		q.Set("flow", flow)
	}
	q.Set("security", security)
	network := node.Transport
	if network == "" {
		network = "tcp"
	}
	q.Set("type", network)

	// security=none 时不写 sni/fp，避免客户端当成 TLS
	if security == "reality" && p.PublicKey != "" {
		q.Set("pbk", p.PublicKey)
		if p.ShortID != "" {
			q.Set("sid", p.ShortID)
		}
		sni := p.HandshakeHost
		if sni == "" {
			sni = p.ServerName
		}
		if sni != "" {
			q.Set("sni", sni)
		}
		q.Set("fp", "chrome")
	} else if security == "tls" {
		sni := p.ServerName
		if sni == "" {
			sni = node.Address
		}
		q.Set("sni", sni)
		q.Set("fp", "chrome")
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
	if p.AllowInsecure {
		q.Set("insecure", "1")
	}
	if p.UpMbps > 0 {
		q.Set("upmbps", fmt.Sprintf("%d", p.UpMbps))
	}
	if p.DownMbps > 0 {
		q.Set("downmbps", fmt.Sprintf("%d", p.DownMbps))
	}

	if p.ObfsEnabled && p.ObfsPassword != "" {
		obfsType := p.ObfsType
		if obfsType == "" {
			obfsType = "salamander"
		}
		q.Set("obfs", obfsType)
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
	if p.AllowInsecure {
		q.Set("allow_insecure", "1")
	}

	return fmt.Sprintf("tuic://%s:%s@%s:%d?%s#%s",
		user.UUID, password, node.Address, node.Port, q.Encode(), url.QueryEscape(node.Name))
}

// ==================== Surge 格式构建 ====================

func buildSurgeVLESS(node model.Node, user model.User, p NodeParams) string {
	sni := p.ServerName
	if sni == "" {
		sni = node.Address
	}

	hasTLS := node.Security == "tls" || node.Security == "reality"
	line := fmt.Sprintf("%s = vless, %s, %d, uuid=%s, tls=%v, sni=%s",
		node.Name, node.Address, node.Port, user.UUID, hasTLS, sni)

	if p.AllowInsecure && hasTLS {
		line += ", skip-cert-verify=true"
	}
	if node.FlowControl != "" && node.FlowControl != "none" {
		line += fmt.Sprintf(", flow=%s", node.FlowControl)
	}
	if node.Security == "reality" && p.PublicKey != "" {
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

	if p.AllowInsecure {
		line += ", skip-cert-verify=true"
	}
	if p.UpMbps > 0 {
		line += fmt.Sprintf(", upload-bandwidth=%d Mbps", p.UpMbps)
	}
	if p.DownMbps > 0 {
		line += fmt.Sprintf(", download-bandwidth=%d Mbps", p.DownMbps)
	}
	if p.ObfsEnabled && p.ObfsPassword != "" {
		obfsType := p.ObfsType
		if obfsType == "" {
			obfsType = "salamander"
		}
		line += fmt.Sprintf(", obfs=%s, obfs-password=%s", obfsType, p.ObfsPassword)
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

	line := fmt.Sprintf("%s = tuic, %s, %d, username=%s, password=%s, sni=%s",
		node.Name, node.Address, node.Port, user.UUID, password, sni)

	if p.AllowInsecure {
		line += ", skip-cert-verify=true"
	}
	if p.CongestionCtrl != "" {
		line += fmt.Sprintf(", congestion-controller=%s", p.CongestionCtrl)
	}

	return line
}
