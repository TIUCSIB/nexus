# Nexus

Go + Vue 3 + sing-box 构建的轻量级代理节点管理面板。面向个人和小圈子使用，无公开注册、无支付系统。

## 功能特性

- **多协议支持** — VLESS（Reality）、Hysteria2、TUIC，配置自动生成
- **多格式订阅** — sing-box JSON、Clash YAML、通用 Base64
- **自定义订阅路径** — 默认为 /s/，可在系统设置中修改
- **设备限制** — Agent 自动检测并关闭超出限制的连接
- **流量统计与限额控制** — 实时采集、增量上报、超额拦截
- **节点在线状态实时监控** — 心跳检测、离线告警
- **流量重置策略** — 不重置 / 每月重置 / 按用户周期重置 / 每年重置
- **订阅信息节点** — 订阅列表末尾显示套餐到期时间和剩余流量
- **强制 HTTPS** — 可选开启，HTTP 请求自动跳转
- **深色主题管理界面** — shadcn-vue 组件，响应式布局

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM / SQLite |
| 前端 | Vue 3 / TypeScript / Pinia / shadcn-vue / Tailwind CSS |
| 代理内核 | sing-box |
| 面板-节点通信 | REST API（server_token + node_id 鉴权） |

## 快速开始

### 环境要求

- Go 1.22+（编译）
- Node.js 18+（前端开发）
- Linux amd64/arm64（部署）

### 一键安装（推荐）

**安装面板（在面板服务器上执行）：**

```bash
# 默认端口 6100
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/install-panel.sh)

# 自定义端口
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/install-panel.sh) --port 8080
```

安装完成后访问 http://your-server-ip:6100，默认账号：
- 邮箱：admin@nexus.com
- 密码：12345678

**安装 Agent（在节点服务器上执行）：**

```bash
# 需要先在面板创建节点获得 node_id，并从系统设置获取 server_token
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/install-agent.sh) \
  --panel https://your-panel.com \
  --node-id 1 \
  --token YOUR_SERVER_TOKEN
```

### 编译

```bash
make build          # 编译面板
make build-agent    # 编译 Agent
make all            # 全部编译
```

### 交叉编译（在 Windows 上编译 Linux 版本）

```powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o bin/nexus-linux-amd64 ./cmd/nexus/

cd agent
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o ../bin/nexus-agent-linux-amd64 ./cmd/agent/
```

### 手动部署

```bash
# 1. 上传文件到服务器
scp bin/nexus-linux-amd64 root@server:/opt/nexus/nexus
scp bin/web-dist.zip root@server:/opt/nexus/
scp config.yaml root@server:/opt/nexus/

# 2. 在服务器上解压前端资源
ssh root@server
mkdir -p /opt/nexus/web/dist /opt/nexus/data
cd /opt/nexus && unzip web-dist.zip -d web/dist

# 3. 创建 systemd 服务
cat > /etc/systemd/system/nexus.service << EOF
[Unit]
Description=Nexus Panel
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/nexus
ExecStart=/opt/nexus/nexus -config /opt/nexus/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 4. 启动
systemctl daemon-reload
systemctl enable nexus
systemctl start nexus

# 5. 创建管理员
/opt/nexus/nexus -config /opt/nexus/config.yaml -admin-email admin@example.com -admin-pass yourpassword
```

### 运行 Agent

Agent 部署在代理节点服务器上，需要先在面板创建节点并配置 agent.yaml：

```yaml
# agent.yaml
panel:
  url: "http://your-panel.com:6100"
  token: "面板系统设置中的 server_token"

nodes:
  - node_id: 1
    address: "0.0.0.0"
    singbox:
      binary_path: "sing-box"
      config_path: "singbox-1.json"
      working_dir: "."
      stats_port: 9091
```

```bash
# 启动 Agent
./nexus-agent -config agent.yaml
```

### 前端开发

```bash
cd web
npm install
npm run dev      # 开发服务器 http://localhost:5173
npm run build    # 生产构建
```

## 节点部署流程

```
1. 在面板中创建节点，记下 node_id
2. 在系统设置中获取 server_token（首次启动自动生成）
3. 在节点服务器上配置 agent.yaml（填入 node_id + server_token）
4. 启动 Agent，自动拉取配置、启动 sing-box
5. 面板中节点状态变为在线，用户可使用订阅链接
```

## 管理命令

```bash
# 面板管理
systemctl start nexus
systemctl stop nexus
systemctl restart nexus
systemctl status nexus
journalctl -u nexus -f

# Agent 管理
systemctl start nexus-agent
systemctl stop nexus-agent
systemctl restart nexus-agent
journalctl -u nexus-agent -f
```

## API 概览

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/login | 登录 |
| POST | /api/auth/refresh | 刷新 Token |

### 用户

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/profile | 个人资料 |
| PUT | /api/user/profile | 更新资料 |
| GET | /api/user/subscription | 订阅信息 |

### 节点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/nodes | 节点列表 |
| GET | /api/nodes/:id/status | 节点状态 |

### 订阅（动态路由）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/{sub_path}/{token} | 获取订阅 |

### 管理端

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST/PUT/DELETE | /api/admin/users | 用户管理 |
| GET/POST/PUT/DELETE | /api/admin/plans | 套餐管理 |
| GET/POST/PUT/DELETE | /api/admin/nodes | 节点管理 |
| GET/POST/PUT/DELETE | /api/admin/groups | 权限组管理 |
| GET/POST/PUT/DELETE | /api/admin/routes | 路由管理 |
| GET/PUT | /api/admin/settings | 系统设置 |
| GET | /api/admin/stats/overview | 统计概览 |
| GET | /api/admin/stats/traffic | 流量统计 |
| GET | /api/admin/online-ips | 在线 IP |
| GET | /api/admin/traffic-logs | 流量日志 |

### Agent 通信（server_token 鉴权）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/internal/agent/:node_id/heartbeat | 心跳上报 |
| GET | /api/internal/agent/:node_id/config | 拉取 sing-box 配置 |
| POST | /api/internal/agent/:node_id/traffic | 上报流量增量 |
| POST | /api/internal/agent/:node_id/alive | 上报在线 IP |
| GET | /api/internal/agent/alivelist | 获取在线用户数 |
| GET | /api/internal/agent/devicelimit | 获取设备限制 |

## 项目结构

```
Nexus/
  cmd/nexus/                  # 面板入口
  internal/
    config/                   # 配置加载
    database/                 # 数据库初始化和迁移
    model/                    # 数据模型（GORM）
    http/
      handler/                # HTTP 处理器
      middleware/              # JWT / Admin 中间件
      router/                 # 路由定义
    service/                  # 业务逻辑层
      config_generator.go     # sing-box 配置生成
    subscription/             # 订阅格式生成（singbox/clash/universal）
    pkg/                      # 工具包（JWT、密码、加密）
  agent/                      # 节点代理
    cmd/agent/                # Agent 入口
    internal/
      config/                 # Agent 配置（YAML）
      httpclient/             # 面板通信客户端
      proxy/                  # sing-box 进程管理
      collector/              # 流量采集
      devicelimit/            # 设备限制执行器
  web/                        # 前端（Vue 3）
    src/
      api/                    # API 调用模块
      views/                  # 页面组件
      components/ui/          # shadcn-vue 组件
      stores/                 # Pinia 状态管理
      types/                  # TypeScript 类型定义
      router/                 # 路由配置
  config.yaml                 # 面板配置文件
  Makefile                    # 构建脚本
```

## 配置参考

### 面板 config.yaml

```yaml
app:
  name: "Nexus"
  debug: false
  secret_key: "auto-generated"

server:
  host: "0.0.0.0"
  port: 6100

database:
  driver: "sqlite"
  dsn: "data/nexus.db"

jwt:
  secret: "auto-generated"
  expire_hours: 72

node:
  heartbeat_interval: 30
  offline_timeout: 90

subscription:
  traffic_reset_days: 30
  plan_sort: 0
```

### Agent agent.yaml

```yaml
panel:
  url: "http://your-panel.com:6100"
  token: "server_token"           # 面板系统设置中的全局 server_token

nodes:
  - node_id: 1                    # 面板中的节点 ID
    address: "0.0.0.0"
    singbox:
      binary_path: "sing-box"
      config_path: "singbox-1.json"
      working_dir: "."
      stats_port: 9091            # sing-box 统计 API 端口
```

## 测试账号

- 邮箱：admin@nexus.com
- 密码：12345678

## 常见问题

### 中文乱码
原因：PowerShell Out-File 会加 BOM 头
解决：始终用 [System.IO.File]::WriteAllText() 写入文件，指定 UTF-8 无 BOM

### 前端路由冲突
原因：Vite 代理规则和前端路由冲突
解决：所有后端 API 路径必须以 /api/ 开头

### 登录失败
原因：JWT token 过期或格式错误
解决：清除 localStorage 重新登录

### Agent 连接失败
原因：server_token 不匹配或 node_id 不存在
解决：确认面板系统设置中的 server_token，确保节点 ID 正确

### 后端重启后数据库报错
原因：新增字段没有默认值
解决：确保 model 字段有 gorm:"default:xxx" 标签

## 开发规范

详见 AGENTS.md。

## 许可证

MIT License
