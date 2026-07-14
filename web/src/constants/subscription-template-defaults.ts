export const builtinSubscriptionTemplateDefaults: Record<string, string> = {
  subscribe_template_singbox: `{
  "dns": {
    "rules": [
      {
        "outbound": [
          "any"
        ],
        "server": "local"
      },
      {
        "clash_mode": "global",
        "server": "remote"
      },
      {
        "clash_mode": "direct",
        "server": "local"
      },
      {
        "rule_set": [
          "geosite-cn"
        ],
        "server": "local"
      }
    ],
    "servers": [
      {
        "address": "https://1.1.1.1/dns-query",
        "detour": "节点选择",
        "tag": "remote"
      },
      {
        "address": "https://223.5.5.5/dns-query",
        "detour": "direct",
        "tag": "local"
      },
      {
        "address": "rcode://success",
        "tag": "block"
      }
    ],
    "strategy": "prefer_ipv4"
  },
  "experimental": {
    "cache_file": {
      "enabled": true,
      "path": "cache.db",
      "cache_id": "cache_db",
      "store_fakeip": true
    }
  },
  "inbounds": [
    {
      "auto_route": true,
      "domain_strategy": "prefer_ipv4",
      "endpoint_independent_nat": true,
      "address": [
        "172.19.0.1/30",
        "2001:0470:f9da:fdfa::1/64"
      ],
      "mtu": 9000,
      "sniff": true,
      "sniff_override_destination": true,
      "stack": "system",
      "strict_route": true,
      "type": "tun"
    },
    {
      "domain_strategy": "prefer_ipv4",
      "listen": "127.0.0.1",
      "listen_port": 2333,
      "sniff": true,
      "sniff_override_destination": true,
      "tag": "socks-in",
      "type": "socks",
      "users": []
    },
    {
      "domain_strategy": "prefer_ipv4",
      "listen": "127.0.0.1",
      "listen_port": 2334,
      "sniff": true,
      "sniff_override_destination": true,
      "tag": "mixed-in",
      "type": "mixed",
      "users": []
    }
  ],
  "outbounds": [
    {
      "tag": "节点选择",
      "type": "selector",
      "default": "自动选择",
      "outbounds": [
        "自动选择",
        "$auto_outbounds"
      ]
    },
    {
      "tag": "direct",
      "type": "direct"
    },
    {
      "tag": "block",
      "type": "block"
    },
    {
      "tag": "dns-out",
      "type": "dns"
    },
    {
      "tag": "自动选择",
      "type": "urltest",
      "outbounds": [
        "$auto_outbounds"
      ]
    }
  ],
  "route": {
    "auto_detect_interface": true,
    "rules": [
      {
        "outbound": "dns-out",
        "protocol": "dns"
      },
      {
        "clash_mode": "direct",
        "outbound": "direct"
      },
      {
        "clash_mode": "global",
        "outbound": "节点选择"
      },
      {
        "ip_is_private": true,
        "outbound": "direct"
      },
      {
        "rule_set": [
          "geosite-cn",
          "geoip-cn"
        ],
        "outbound": "direct"
      }
    ],
    "rule_set": [
      {
        "tag": "geosite-cn",
        "type": "remote",
        "format": "binary",
        "url": "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-cn.srs",
        "download_detour": "自动选择"
      },
      {
        "tag": "geoip-cn",
        "type": "remote",
        "format": "binary",
        "url": "https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set/geoip-cn.srs",
        "download_detour": "自动选择"
      }
    ]
  }
}`,
  subscribe_template_clash: `mixed-port: 7890
allow-lan: true
bind-address: "*"
mode: rule
log-level: info
external-controller: 127.0.0.1:9090
unified-delay: true
tcp-concurrent: true

dns:
  enable: true
  ipv6: false
  default-nameserver:
    - 223.5.5.5
    - 119.29.29.29
  enhanced-mode: fake-ip
  fake-ip-range: 198.18.0.1/16
  use-hosts: true
  nameserver-policy:
    "+.google.com": "https://dns.cloudflare.com/dns-query"
    "+.googleapis.com": "https://dns.cloudflare.com/dns-query"
    "+.googleapis.cn": "https://dns.cloudflare.com/dns-query"
    "+.googlevideo.com": "https://dns.cloudflare.com/dns-query"
    "+.gstatic.com": "https://dns.cloudflare.com/dns-query"
    "+.youtube.com": "https://dns.cloudflare.com/dns-query"
    "+.youtu.be": "https://dns.cloudflare.com/dns-query"
    "+.facebook.com": "https://dns.cloudflare.com/dns-query"
    "+.twitter.com": "https://dns.cloudflare.com/dns-query"
    "+.x.com": "https://dns.cloudflare.com/dns-query"
    "+.github.com": "https://dns.cloudflare.com/dns-query"
    "+.githubusercontent.com": "https://dns.cloudflare.com/dns-query"
    "+.openai.com": "https://dns.cloudflare.com/dns-query"
    "+.chatgpt.com": "https://dns.cloudflare.com/dns-query"
    "+.anthropic.com": "https://dns.cloudflare.com/dns-query"
  nameserver:
    - https://doh.pub/dns-query
    - https://dns.alidns.com/dns-query
    - tls://dot.pub:853
    - tls://dns.alidns.com:853
  fallback:
    - https://dns.cloudflare.com/dns-query
    - https://dns.google/dns-query
    - tls://1.1.1.1:853
    - tls://8.8.8.8:853
  fallback-filter:
    geoip: true
    geoip-code: CN
    ipcidr:
      - 0.0.0.0/8
      - 10.0.0.0/8
      - 100.64.0.0/10
      - 127.0.0.0/8
      - 169.254.0.0/16
      - 172.16.0.0/12
      - 192.168.0.0/16
      - 224.0.0.0/4
      - 240.0.0.0/4
    domain:
      - "+.google.com"
      - "+.facebook.com"
      - "+.youtube.com"
      - "+.githubusercontent.com"
      - "+.googlevideo.com"
      - "+.googleapis.cn"
  fake-ip-filter:
    - "*.lan"
    - "*.local"
    - "*.localhost"
    - "*.test"
    - localhost.ptlogin2.qq.com
    - "+.stun.*.*"
    - "+.stun.*.*.*"
    - "+.stun.*.*.*.*"
    - lens.l.google.com
    - "*.srv.nintendo.net"
    - "+.stun.playstation.net"
    - xbox.*.*.microsoft.com
    - "*.*.xboxlive.com"
    - "+.msftncsi.com"
    - "+.msftconnecttest.com"

proxies:

proxy-groups:
  - { name: "$app_name", type: select, proxies: ["自动选择", "故障转移", "$auto_proxy", "DIRECT"] }
  - { name: "自动选择", type: url-test, proxies: ["$auto_proxy"], url: "http://www.gstatic.com/generate_204", interval: 300, tolerance: 50 }
  - { name: "故障转移", type: fallback, proxies: ["$auto_proxy"], url: "http://www.gstatic.com/generate_204", interval: 300 }

rules:
  - DOMAIN-SUFFIX,services.googleapis.cn,$app_name
  - DOMAIN-SUFFIX,xn--ngstr-lra8j.com,$app_name
  - DOMAIN-SUFFIX,apps.apple.com,$app_name
  - DOMAIN-SUFFIX,blobstore.apple.com,$app_name
  - DOMAIN,safebrowsing.urlsec.qq.com,DIRECT
  - DOMAIN-SUFFIX,apple.com,DIRECT
  - DOMAIN-SUFFIX,apple-cloudkit.com,DIRECT
  - DOMAIN-SUFFIX,icloud.com,DIRECT
  - DOMAIN-SUFFIX,icloud-content.com,DIRECT
  - DOMAIN-SUFFIX,mzstatic.com,DIRECT
  - DOMAIN-SUFFIX,aaplimg.com,DIRECT
  - DOMAIN-SUFFIX,cdn-apple.com,DIRECT
  - DOMAIN-SUFFIX,akadns.net,DIRECT
  - DOMAIN-KEYWORD,alicdn,DIRECT
  - DOMAIN-KEYWORD,alipay,DIRECT
  - DOMAIN-KEYWORD,taobao,DIRECT
  - DOMAIN-KEYWORD,baidu,DIRECT
  - DOMAIN-KEYWORD,tencent,DIRECT
  - DOMAIN-SUFFIX,qq.com,DIRECT
  - DOMAIN-SUFFIX,gtimg.com,DIRECT
  - DOMAIN-SUFFIX,qcloud.com,DIRECT
  - DOMAIN-SUFFIX,tenpay.com,DIRECT
  - DOMAIN-SUFFIX,tmall.com,DIRECT
  - DOMAIN-SUFFIX,jd.com,DIRECT
  - DOMAIN-SUFFIX,iqiyi.com,DIRECT
  - DOMAIN-SUFFIX,youku.com,DIRECT
  - DOMAIN-SUFFIX,zhihu.com,DIRECT
  - DOMAIN-SUFFIX,zhimg.com,DIRECT
  - DOMAIN-SUFFIX,163.com,DIRECT
  - DOMAIN-SUFFIX,126.com,DIRECT
  - DOMAIN-SUFFIX,weibo.com,DIRECT
  - DOMAIN-SUFFIX,ximalaya.com,DIRECT
  - DOMAIN-SUFFIX,csdn.net,DIRECT
  - DOMAIN-SUFFIX,gitee.com,DIRECT
  - DOMAIN-SUFFIX,jianshu.com,DIRECT
  - DOMAIN-SUFFIX,cnblogs.com,DIRECT
  - DOMAIN-SUFFIX,oschina.net,DIRECT
  - DOMAIN-SUFFIX,ele.me,DIRECT
  - DOMAIN-SUFFIX,ctrip.com,DIRECT
  - DOMAIN-SUFFIX,dianping.com,DIRECT
  - DOMAIN-SUFFIX,amap.com,DIRECT
  - DOMAIN-SUFFIX,autonavi.com,DIRECT
  - DOMAIN-SUFFIX,mi.com,DIRECT
  - DOMAIN-SUFFIX,miui.com,DIRECT
  - DOMAIN-SUFFIX,ifeng.com,DIRECT
  - DOMAIN-SUFFIX,youdao.com,DIRECT
  - DOMAIN-SUFFIX,iciba.com,DIRECT
  - DOMAIN-SUFFIX,xunlei.com,DIRECT
  - DOMAIN-SUFFIX,smzdm.com,DIRECT
  - DOMAIN-SUFFIX,sspai.com,DIRECT
  - DOMAIN-SUFFIX,36kr.com,DIRECT
  - DOMAIN-SUFFIX,speedtest.net,DIRECT
  - DOMAIN-SUFFIX,microsoft.com,DIRECT
  - DOMAIN-SUFFIX,microsoftonline.com,DIRECT
  - DOMAIN-SUFFIX,office.com,DIRECT
  - DOMAIN-SUFFIX,office365.com,DIRECT
  - DOMAIN-SUFFIX,windows.com,DIRECT
  - DOMAIN-SUFFIX,windowsupdate.com,DIRECT
  - DOMAIN-SUFFIX,live.com,DIRECT
  - DOMAIN-SUFFIX,msn.com,DIRECT
  - DOMAIN-SUFFIX,cn,DIRECT
  - DOMAIN-KEYWORD,-cn,DIRECT
  - DOMAIN-KEYWORD,admarvel,REJECT
  - DOMAIN-KEYWORD,admaster,REJECT
  - DOMAIN-KEYWORD,adsage,REJECT
  - DOMAIN-KEYWORD,adsmogo,REJECT
  - DOMAIN-KEYWORD,adsrvmedia,REJECT
  - DOMAIN-KEYWORD,adwords,REJECT
  - DOMAIN-KEYWORD,adservice,REJECT
  - DOMAIN-KEYWORD,domob,REJECT
  - DOMAIN-KEYWORD,duomeng,REJECT
  - DOMAIN-KEYWORD,dwtrack,REJECT
  - DOMAIN-KEYWORD,guanggao,REJECT
  - DOMAIN-KEYWORD,lianmeng,REJECT
  - DOMAIN-KEYWORD,omgmta,REJECT
  - DOMAIN-KEYWORD,openx,REJECT
  - DOMAIN-KEYWORD,partnerad,REJECT
  - DOMAIN-KEYWORD,supersonicads,REJECT
  - DOMAIN-KEYWORD,umeng,REJECT
  - DOMAIN-KEYWORD,zjtoolbar,REJECT
  - DOMAIN-SUFFIX,appsflyer.com,REJECT
  - DOMAIN-SUFFIX,doubleclick.net,REJECT
  - DOMAIN-SUFFIX,mmstat.com,REJECT
  - DOMAIN-KEYWORD,google,$app_name
  - DOMAIN-KEYWORD,gmail,$app_name
  - DOMAIN-KEYWORD,youtube,$app_name
  - DOMAIN-KEYWORD,facebook,$app_name
  - DOMAIN-KEYWORD,twitter,$app_name
  - DOMAIN-KEYWORD,instagram,$app_name
  - DOMAIN-KEYWORD,whatsapp,$app_name
  - DOMAIN-KEYWORD,telegram,$app_name
  - DOMAIN-KEYWORD,github,$app_name
  - DOMAIN-KEYWORD,blogspot,$app_name
  - DOMAIN-KEYWORD,dropbox,$app_name
  - DOMAIN-KEYWORD,wikipedia,$app_name
  - DOMAIN-KEYWORD,pinterest,$app_name
  - DOMAIN-KEYWORD,discord,$app_name
  - DOMAIN-KEYWORD,openai,$app_name
  - DOMAIN-KEYWORD,anthropic,$app_name
  - DOMAIN-KEYWORD,netflix,$app_name
  - DOMAIN-KEYWORD,spotify,$app_name
  - DOMAIN-KEYWORD,amazon,$app_name
  - DOMAIN-SUFFIX,t.co,$app_name
  - DOMAIN-SUFFIX,x.com,$app_name
  - DOMAIN-SUFFIX,twimg.com,$app_name
  - DOMAIN-SUFFIX,fb.me,$app_name
  - DOMAIN-SUFFIX,fbcdn.net,$app_name
  - DOMAIN-SUFFIX,youtu.be,$app_name
  - DOMAIN-SUFFIX,ytimg.com,$app_name
  - DOMAIN-SUFFIX,gstatic.com,$app_name
  - DOMAIN-SUFFIX,ggpht.com,$app_name
  - DOMAIN-SUFFIX,googlevideo.com,$app_name
  - DOMAIN-SUFFIX,v2ex.com,$app_name
  - DOMAIN-SUFFIX,medium.com,$app_name
  - DOMAIN-SUFFIX,reddit.com,$app_name
  - DOMAIN-SUFFIX,redd.it,$app_name
  - DOMAIN-SUFFIX,imgur.com,$app_name
  - DOMAIN-SUFFIX,pixiv.net,$app_name
  - DOMAIN-SUFFIX,nytimes.com,$app_name
  - DOMAIN-SUFFIX,nyt.com,$app_name
  - DOMAIN-SUFFIX,bbc.com,$app_name
  - DOMAIN-SUFFIX,bbc.co.uk,$app_name
  - DOMAIN-SUFFIX,steamcommunity.com,$app_name
  - DOMAIN-SUFFIX,twitch.tv,$app_name
  - DOMAIN-SUFFIX,vimeo.com,$app_name
  - DOMAIN-SUFFIX,tumblr.com,$app_name
  - DOMAIN-SUFFIX,linkedin.com,$app_name
  - DOMAIN-SUFFIX,licdn.com,$app_name
  - DOMAIN-SUFFIX,mega.nz,$app_name
  - DOMAIN-SUFFIX,archive.org,$app_name
  - DOMAIN-SUFFIX,wikimedia.org,$app_name
  - DOMAIN-SUFFIX,soundcloud.com,$app_name
  - IP-CIDR,91.108.4.0/22,$app_name,no-resolve
  - IP-CIDR,91.108.8.0/21,$app_name,no-resolve
  - IP-CIDR,91.108.12.0/22,$app_name,no-resolve
  - IP-CIDR,91.108.16.0/22,$app_name,no-resolve
  - IP-CIDR,91.108.56.0/22,$app_name,no-resolve
  - IP-CIDR,149.154.160.0/20,$app_name,no-resolve
  - IP-CIDR6,2001:67c:4e8::/48,$app_name,no-resolve
  - IP-CIDR6,2001:b28:f23d::/48,$app_name,no-resolve
  - IP-CIDR6,2001:b28:f23f::/48,$app_name,no-resolve
  - GEOIP,CN,DIRECT
  - MATCH,$app_name
`,
  subscribe_template_clashmeta: '',
  subscribe_template_stash: '',
  subscribe_template_surge: `#!MANAGED-CONFIG interval=43200 strict=true
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
`,
  subscribe_template_surfboard: `#!MANAGED-CONFIG interval=43200 strict=true

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
`,
}

builtinSubscriptionTemplateDefaults.subscribe_template_clashmeta = builtinSubscriptionTemplateDefaults.subscribe_template_clash
builtinSubscriptionTemplateDefaults.subscribe_template_stash = builtinSubscriptionTemplateDefaults.subscribe_template_clash
