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
- **实时 WebSocket 通信** — 面板与 Agent 实时指令下发
- **机器管理模式** — 单 Agent 管理多节点，系统负载监控
- **操作审计日志** — 记录所有管理操作
- **节点/用户流量排行** — 统计流量使用情况
- **FakeDNS + DNS 缓存** — FakeIP 加速
- **安全加固** — 登录锁定 / Token 轮换 / 安全头 / WS Origin 校验

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM / SQLite |
| 前端 | Vue 3 / TypeScript / Pinia / shadcn-vue / Tailwind CSS |
| 代理内核 | sing-box |
| 面板-节点通信 | REST API（server_token + node_id 鉴权） |

## 快速开始

### 📦 预编译二进制（推荐）

从 [GitHub Releases](https://github.com/TIUCSIB/Nexus/releases) 下载最新版本的预编译二进制文件，支持：

- Linux (amd64/arm64)
- Windows (amd64)
- macOS (amd64/arm64)

包含：面板、Agent、NS CLI 和前端静态文件。

### 🚀 一键安装脚本

安装脚本会自动从 GitHub Releases 下载最新版本。

### 环境要求

- Go 1.22+（仅编译时需要）
- Node.js 20+（仅前端开发时需要）
- Linux amd64/arm64（部署运行）

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

### 🔄 管理面板和节点

使用管理脚本：

```bash
# 下载并运行管理菜单
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/nexus.sh)

# 或直接使用 systemctl
systemctl start/stop/restart nexus        # 面板
systemctl start/stop/restart nexus-agent  # Agent
journalctl -u nexus -f                    # 查看日志

# Agent 管理（安装后可用）
ns status              # 查看状态
ns list                # 列出节点
ns service restart     # 重启服务
```

### 🐳 Docker 部署

```bash
# 构建并启动
docker compose up -d

# 查看日志
docker compose logs -f

# 停止
docker compose down

# 更新
docker compose build --no-cache && docker compose up -d
```

默认映射端口 8080，数据持久化在 `nexus-data` volume 中。如需自定义配置，修改 `config.yaml` 后重启。

## 开发指南

### 本地编译

```bash
make build          # 编译面板
make build-agent    # 编译 Agent
make all            # 全部编译
```

### 前端开发

```bash
cd web
npm install
npm run dev         # 开发服务器
npm run build       # 生产构建
```

### 🚀 发布新版本

使用快速发布脚本：

```bash
# 自动提交、打标签、触发 GitHub Actions 构建
./release.sh v1.0.0
```

或手动发布：

```bash
# 1. 提交所有更改
git add .
git commit -m "feat: 新功能"
git push

# 2. 创建并推送标签
git tag v1.0.0
git push origin v1.0.0

# 3. GitHub Actions 会自动构建并发布到 Releases
```

查看构建进度：[GitHub Actions](https://github.com/TIUCSIB/Nexus/actions)

发布后，安装脚本会自动使用最新版本。详见 [.github/WORKFLOWS.md](.github/WORKFLOWS.md)

### 交叉编译（手动）

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
| GET/POST/PUT/DELETE | /api/admin/custom-outbounds | 自定义出站管理 |
| GET/PUT | /api/admin/nodes/:id/outbounds | 节点出站绑定 |
| GET/PUT | /api/admin/settings | 系统设置 |
| GET | /api/admin/stats/overview | 统计概览 |
| GET | /api/admin/stats/traffic | 流量统计（按天） |
| GET | /api/admin/online-ips | 在线 IP |
| GET | /api/admin/traffic-logs | 流量日志 |
| POST | /api/admin/users/:id/reset-uuid | 重置用户 UUID |
| POST | /api/admin/users/:id/reset-traffic | 重置用户流量 |
| POST | /api/admin/nodes/:id/restart | 重启节点 |
| POST | /api/admin/nodes/:id/reset-traffic | 重置节点流量 |

### 统计与分析

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/stats/overview | 统计概览 |
| GET | /api/admin/stats/traffic | 流量统计（按天） |
| GET | /api/admin/stats/node-ranking | 节点流量排行 |
| GET | /api/admin/stats/user-ranking | 用户流量排行 |
| GET | /api/admin/stats/system | 系统状态 |

### 审计与流量重置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/audit-logs | 审计日志 |
| GET | /api/admin/traffic-reset/users | 流量重置用户列表 |
| POST | /api/admin/traffic-reset/manual | 手动流量重置 |
| GET | /api/admin/traffic-reset/stats | 流量重置统计 |
| POST | /api/admin/users/batch-operation | 批量操作 |

### Agent 通信（server_token 鉴权）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/internal/agent/handshake | 握手协商（获取推拉间隔） |
| POST | /api/internal/agent/:node_id/heartbeat | 心跳上报 |
| GET | /api/internal/agent/:node_id/config | 拉取节点配置（参数化/JSON 模式） |
| GET | /api/internal/agent/:node_id/users | 拉取活跃用户列表（ETag 缓存） |
| POST | /api/internal/agent/:node_id/report | 合并上报（流量+在线IP+状态） |
| POST | /api/internal/agent/:node_id/traffic | 上报流量增量 |
| POST | /api/internal/agent/:node_id/alive | 上报在线 IP |
| GET | /api/internal/agent/alivelist | 获取在线用户数 |

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
    cmd/ns/                   # ns CLI 工具（bind/list/service）
    internal/
      config/                 # Agent 配置（YAML，支持 Nexus/Xboard 格式）
      httpclient/             # 面板通信客户端
      proxy/                  # sing-box 进程管理
      collector/              # 流量采集
      kernel/                 # sing-box 配置生成器
      cert/                   # 证书自动化（ACME/自签名/文件）
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

两种配置格式都支持：

**Nexus 原生格式：**
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

**Xboard 兼容格式：**
```yaml
panel:
  url: "http://your-panel.com:6100"
  token: "server_token"

kernel:
  type: "singbox"

instances:
  - panel:
      node_id: 1
    machine:
      machine_id: 1
      token: "server_token"
    singbox:
      binary_path: "sing-box"
      config_path: "singbox-1.json"
      working_dir: "."
      stats_port: 9091
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
