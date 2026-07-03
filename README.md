# Nexus

Go + Vue 3 + sing-box 构建的轻量级代理节点管理面板。面向个人和小圈子使用，无公开注册、无支付系统。

## 功能特性

- 多协议支持：VLESS（Reality）、Hysteria2、TUIC
- 多格式订阅：sing-box JSON、Clash YAML、通用 Base64
- 自定义订阅路径（默认 /s/）
- 设备限制：自动检测并关闭超出限制的连接
- 流量统计与限额控制
- 节点在线状态实时监控
- 暗色主题管理界面
- 一键安装脚本（类 V2bX 风格）

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM / SQLite |
| 前端 | Vue 3 / TypeScript / Pinia / shadcn-vue / Tailwind CSS |
| 代理内核 | sing-box |
| 通信方式 | REST API |

## 快速开始

### 一键安装（推荐）

**安装面板（在面板服务器上执行）：**

```bash
# 默认端口 6100
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/install-panel.sh)

# 自定义端口
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/install-panel.sh) --port 8080
```

安装完成后访问 `http://your-server-ip:6100`，默认账号：
- 邮箱：`admin@nexus.com`
- 密码：`12345678`

**安装 Agent（在节点服务器上执行）：**

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/install-agent.sh) \
  --panel https://your-panel.com \
  --token YOUR_REGISTER_TOKEN \
  --name my-node
```

### 环境要求

- Go 1.22+（编译）
- Node.js 18+（前端开发）
- Linux amd64/arm64（部署）

### 编译

```bash
make build          # 编译面板
make build-agent    # 编译 Agent
make all            # 全部编译
```

### 交叉编译（在 Windows 上编译 Linux 版本）

```bash
set GOOS=linux
set GOARCH=amd64
go build -o bin/nexus-linux-amd64 ./cmd/nexus/

cd agent
set GOOS=linux
set GOARCH=amd64
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
cat > /etc/systemd/system/nexus.service << 'EOF'
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

Agent 部署在代理节点服务器上，负责管理 sing-box 进程。

**方式一：命令行参数（推荐）**

```bash
./nexus-agent --panel https://your-panel.com --token YOUR_TOKEN --name my-node
```

**方式二：环境变量**

```bash
export NEXUS_PANEL_URL=https://your-panel.com
export NEXUS_TOKEN=YOUR_TOKEN
export NEXUS_NODE_NAME=my-node
./nexus-agent
```

**方式三：配置文件**

```bash
cp agent/config.yml.example agent.yaml
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
1. 在面板创建节点，复制 register_token
2. 在节点服务器运行安装命令
3. Agent 自动注册、拉取配置、启动 sing-box
4. 节点出现在面板中，用户可以使用订阅链接
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
- `POST /api/auth/login` - 登录
- `POST /api/auth/refresh` - 刷新 Token

### 订阅（动态路由）
- `GET /api/{sub_path}/{token}` - 获取订阅

### 用户
- `GET /api/user/profile` - 个人资料
- `GET /api/user/subscription` - 订阅信息

### 管理
- `/api/admin/users` - 用户管理
- `/api/admin/plans` - 套餐管理
- `/api/admin/nodes` - 节点管理
- `/api/admin/settings` - 系统设置

### Agent 通信
- `POST /api/internal/agent/register` - 注册
- `POST /api/internal/agent/heartbeat` - 心跳
- `GET /api/internal/agent/config` - 拉取配置
- `POST /api/internal/agent/traffic` - 上报流量
- `GET /api/internal/agent/devicelimit` - 获取设备限制

## 项目结构

```
Nexus/
  cmd/nexus/                  # 面板入口
  internal/
    config/                   # 配置加载
    database/                 # 数据库初始化
    model/                    # 数据模型（GORM）
    http/
      handler/                # HTTP 处理器
      middleware/              # JWT / Admin 中间件
      router/                 # 路由定义
    service/                  # 业务逻辑
    subscription/             # 订阅格式生成
    pkg/                      # 工具包
  agent/
    cmd/agent/                # Agent 入口
    internal/
      config/                 # Agent 配置（YAML/CLI/环境变量）
      httpclient/             # 面板通信
      proxy/                  # sing-box 管理
      collector/              # 流量采集
      devicelimit/            # 设备限制执行
  web/                        # 前端（Vue 3）
  scripts/
    install-panel.sh          # 面板一键安装
    install-agent.sh          # Agent 一键安装
    nexus.sh                  # 管理脚本
  config.yaml                 # 面板配置
  Makefile                    # 构建脚本
```

## 配置文件

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

### Agent 命令行参数

| 参数 | 环境变量 | 说明 |
|------|----------|------|
| --panel | NEXUS_PANEL_URL | 面板地址（必填） |
| --token | NEXUS_TOKEN | 注册令牌（必填） |
| --name | NEXUS_NODE_NAME | 节点名称（默认 node-1） |
| --port | - | 统计端口（默认 9090） |

## 开发

详细开发规范请参看 AGENTS.md

## 许可证

MIT License
