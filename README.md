# Nexus

Go + Vue 3 + sing-box 构建的轻量级代理节点管理面板。
面向个人和小圈子使用，无公开注册、无支付系统。

## 功能特性

- 多协议支持：VLESS（Reality）、Hysteria2、TUIC
- 多格式订阅：sing-box JSON、Clash YAML、通用 Base64
- 自定义订阅路径（默认 /s/）
- 设备限制：自动检测并关闭超出限制的连接
- 流量统计与限额控制
- 节点在线状态实时监控
- 暗色主题管理界面

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM / SQLite |
| 前端 | Vue 3 / TypeScript / Pinia / shadcn-vue / Tailwind CSS |
| 代理内核 | sing-box |
| 通信方式 | REST API |

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+（前端开发）
- GCC（编译 SQLite 驱动）

### 编译

```bash
make build          # 编译面板
make build-agent    # 编译 Agent
make all            # 全部编译
```

### 运行面板

```bash
./bin/nexus.exe
```

访问 http://localhost:8080 ，使用测试账号登录：
- 邮箱：admin@nexus.com
- 密码：12345678

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

**Linux 一键安装**

```bash
bash install.sh --panel https://your-panel.com --token YOUR_TOKEN --name my-node
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
2. 在服务器运行：./nexus-agent --panel URL --token TOKEN --name 节点名
3. Agent 自动注册、拉取配置、启动 sing-box
4. 节点出现在面板中，用户可以使用订阅链接
```

## Docker 部署

```bash
# 面板
docker-compose up -d

# Agent
docker run -d --network=host \\
  -e NEXUS_PANEL_URL=https://your-panel.com \\
  -e NEXUS_TOKEN=YOUR_TOKEN \\
  -e NEXUS_NODE_NAME=my-node \\
  nexus-agent
```

## API 概览

### 认证
- POST /api/auth/login - 登录
- POST /api/auth/refresh - 刷新 Token

### 订阅（动态路径）
- GET /api/{sub_path}/{token} - 获取订阅
- 支持 query 参数：format=singbox/clash/base64

### 用户
- GET /api/user/profile - 个人资料
- GET /api/user/subscription - 订阅信息

### 管理
- /api/admin/users - 用户管理
- /api/admin/plans - 套餐管理
- /api/admin/nodes - 节点管理
- /api/admin/settings - 系统设置
- /api/admin/stats/* - 统计数据

### Agent 通信
- POST /api/internal/agent/register - 注册
- POST /api/internal/agent/heartbeat - 心跳
- GET /api/internal/agent/config - 拉取配置
- POST /api/internal/agent/traffic - 上报流量
- POST /api/internal/agent/alive - 上报在线IP
- GET /api/internal/agent/devicelimit - 获取设备限制

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
    install.sh                # 一键安装脚本
    config.yml.example        # 示例配置
  web/                        # 前端（Vue 3）
  config.yaml                 # 面板配置
  Makefile                    # 构建脚本
```

## 配置文件

### 面板 config.yaml

```yaml
app:
  name: "Nexus"
  debug: true
  secret_key: "change-me"

server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"
  dsn: "data/nexus.db"

jwt:
  secret: "change-me-jwt-secret"
  expire_hours: 72
```

### Agent 命令行参数

| 参数 | 环境变量 | 说明 |
|------|----------|------|
| --panel | NEXUS_PANEL_URL | 面板地址（必填） |
| --token | NEXUS_TOKEN | 注册令牌（必填） |
| --name | NEXUS_NODE_NAME | 节点名称（默认 node-1） |
| --config | NEXUS_CONFIG | 配置文件路径 |

## 开发

详细开发规范请参考 AGENTS.md

## 许可证

MIT License
