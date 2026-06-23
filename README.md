# Nexus 代理面板

Nexus 是一个现代化的代理节点管理面板，支持多协议、多用户和多客户端订阅格式，适合个人和小型团队使用。

## ✨ 功能特性

### 用户管理
- 管理员创建和管理用户账号
- 用户登录、Token 刷新
- 用户资料查看和修改
- 用户状态控制（启用/禁用）

### 套餐管理
- 灵活的流量套餐配置
- 自定义套餐时长、流量限额、价格
- 套餐排序和上下架管理

### 节点管理
- 支持 VLESS（Reality）、Hysteria2、TUIC 三种协议
- 节点状态实时监控（在线/离线）
- 通过 Agent 自动上报节点状态
- 节点配置集中管理

### 订阅系统
- sing-box JSON 格式
- Clash / Clash.Meta YAML 格式
- Surge 配置格式
- Surfboard 配置格式
- V2RayN / Shadowrocket Base64 格式
- Token 认证，无需 JWT

### 流量统计
- 用户流量使用统计
- 流量日志记录
- 流量限额控制

## 🛠 技术栈

| 组件 | 技术 |
|------|------|
| 后端语言 | Go 1.22+ |
| Web 框架 | Gin |
| 数据库 | SQLite（GORM） |
| 认证 | JWT |
| 前端 | Vue 3 + TypeScript |
| 节点通信 | gRPC |
| 配置格式 | YAML |

## 🚀 快速开始

### 环境要求

- Go 1.22 或更高版本
- GCC（用于编译 SQLite 驱动）

### 编译

```bash
# 编译面板
make build

# 编译 Agent
make build-agent

# 全部编译
make all
```

### 运行

```bash
# 首次运行：创建管理员账号
./bin/nexus.exe -init-admin

# 正常启动
./bin/nexus.exe -config config.yaml
```

### 配置文件

编辑 `config.yaml`：

```yaml
app:
  name: "Nexus"
  debug: false                    # 生产环境设为 false
  secret_key: "your-random-key"   # 应用密钥，务必修改

server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"
  dsn: "data/nexus.db"           # SQLite 数据库路径

jwt:
  secret: "your-jwt-secret"      # JWT 密钥，务必修改
  expire_hours: 72               # Token 有效期（小时）

grpc:
  listen: "0.0.0.0:9090"        # Agent 通信端口
  cert_file: ""                  # TLS 证书路径（留空则不加密）
  key_file: ""                   # TLS 密钥路径

node:
  heartbeat_interval: 30         # 节点心跳间隔（秒）
  offline_timeout: 90            # 节点离线超时（秒）

subscription:
  traffic_reset_days: 30         # 流量重置周期（天），0 为不重置
```

## 📡 节点 Agent 配置

Agent 部署在每一台代理节点服务器上，负责管理 sing-box 进程并上报状态。

编辑 `agent/agent.yaml`：

```yaml
server:
  addr: "panel.example.com:9090"  # 面板 gRPC 地址
  token: "your-node-token"        # 节点注册 Token（在面板中获取）

singbox:
  config_path: "/etc/sing-box/config.json"
  binary_path: "/usr/bin/sing-box"
```

启动 Agent：

```bash
./bin/nexus-agent.exe -config agent.yaml
```

## 📖 API 文档

### 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |

### 订阅接口（Token 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/sub/singbox?token=xxx` | sing-box 订阅 |
| GET | `/api/v1/sub/clash?token=xxx` | Clash 订阅 |
| GET | `/api/v1/sub/surge?token=xxx` | Surge 订阅 |
| GET | `/api/v1/sub/surfboard?token=xxx` | Surfboard 订阅 |
| GET | `/api/v1/sub/shadowrocket?token=xxx` | Shadowrocket 订阅 |
| GET | `/api/v1/sub/v2rayn?token=xxx` | V2RayN 订阅 |

### 用户接口（JWT 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/user/profile` | 获取用户资料 |
| PUT | `/api/v1/user/profile` | 更新用户资料 |
| GET | `/api/v1/user/subscription` | 获取订阅信息 |

### 节点接口（JWT 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/nodes` | 获取节点列表 |
| GET | `/api/v1/nodes/:id/status` | 获取节点状态 |

### 管理员接口（JWT + Admin 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/users` | 用户列表 |
| POST | `/api/v1/admin/users` | 创建用户 |
| PUT | `/api/v1/admin/users/:id` | 更新用户 |
| DELETE | `/api/v1/admin/users/:id` | 删除用户 |
| GET | `/api/v1/admin/plans` | 套餐列表 |
| POST | `/api/v1/admin/plans` | 创建套餐 |
| PUT | `/api/v1/admin/plans/:id` | 更新套餐 |
| DELETE | `/api/v1/admin/plans/:id` | 删除套餐 |
| GET | `/api/v1/admin/nodes` | 节点列表 |
| POST | `/api/v1/admin/nodes` | 创建节点 |
| PUT | `/api/v1/admin/nodes/:id` | 更新节点 |
| DELETE | `/api/v1/admin/nodes/:id` | 删除节点 |
| POST | `/api/v1/admin/nodes/:id/restart` | 重启节点 |
| GET | `/api/v1/admin/settings` | 获取设置 |
| PUT | `/api/v1/admin/settings` | 更新设置 |
| GET | `/api/v1/admin/stats/overview` | 概览统计 |
| GET | `/api/v1/admin/stats/traffic` | 流量统计 |

### 通用响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code = 0` 表示成功
- `code = -1` 表示失败，`message` 字段包含错误信息

## 🐳 Docker 部署

### 使用 Docker Compose（推荐）

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f nexus

# 停止服务
docker-compose down
```

### 单独使用 Docker

```bash
# 构建镜像
docker build -t nexus .

# 运行容器
docker run -d \
  --name nexus \
  -p 8080:8080 \
  -p 9090:9090 \
  -v nexus-data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml \
  nexus
```

### Agent Docker 部署

```bash
# 构建 Agent 镜像
cd agent
docker build -t nexus-agent .

# 运行 Agent
docker run -d \
  --name nexus-agent \
  -v $(pwd)/agent.yaml:/app/agent.yaml \
  nexus-agent
```

## 💻 裸机部署

### Linux

```bash
# 安装依赖（Ubuntu/Debian）
sudo apt update
sudo apt install -y gcc musl-dev sqlite3

# 编译
make all

# 创建 systemd 服务
sudo tee /etc/systemd/system/nexus.service <<EOF
[Unit]
Description=Nexus Proxy Panel
After=network.target

[Service]
Type=simple
User=nexus
WorkingDirectory=/opt/nexus
ExecStart=/opt/nexus/nexus -config /opt/nexus/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动
sudo systemctl daemon-reload
sudo systemctl enable nexus
sudo systemctl start nexus
```

### Windows

```powershell
# 编译
.\make.bat build

# 运行（建议使用 NSSM 注册为 Windows 服务）
.\bin\nexus.exe -config config.yaml
```

## 📁 项目结构

```
nexus/
├── cmd/nexus/              # 主程序入口
├── internal/
│   ├── config/             # 配置加载
│   ├── database/           # 数据库初始化
│   ├── http/
│   │   ├── handler/        # HTTP 处理器
│   │   ├── middleware/      # 中间件（JWT、管理员）
│   │   └── router/         # 路由定义
│   ├── model/              # 数据模型
│   ├── pkg/                # 公共库（加密、JWT、流量）
│   ├── proto/              # Protobuf 定义
│   ├── service/            # 业务服务层
│   └── subscription/       # 订阅格式生成器
├── agent/                  # 节点 Agent
│   ├── cmd/agent/          # Agent 入口
│   └── internal/           # Agent 内部实现
├── api/proto/              # Proto 源文件
├── web/                    # 前端代码
├── config.yaml             # 面板配置
├── docker-compose.yml      # Docker Compose
├── Dockerfile              # 面板 Dockerfile
└── Makefile                # 构建脚本
```

## 🔒 安全建议

1. **务必修改** `config.yaml` 中的 `secret_key` 和 `jwt.secret`
2. 生产环境将 `app.debug` 设置为 `false`
3. 建议使用 Nginx 反向代理，并启用 HTTPS
4. gRPC 端口建议配置 TLS 或限制访问 IP
5. 定期备份 `data/nexus.db` 数据库文件

## 📄 许可证

MIT License

Copyright (c) 2024 Nexus

特此免费授予任何获得本软件及相关文档文件（"软件"）副本的人不受限制地处理本软件的权限，
包括但不限于使用、复制、修改、合并、发布、分发、再许可和/或销售本软件副本的权利，
并允许向其提供本软件的人这样做，但须符合以下条件：

上述版权声明和本许可声明应包含在本软件的所有副本或主要部分中。

本软件按"原样"提供，不附带任何明示或暗示的担保，包括但不限于对适销性、
特定用途适用性和非侵权性的担保。在任何情况下，作者或版权持有人均不对因本软件或
本软件的使用或其他交易而产生的任何索赔、损害或其他责任负责。
