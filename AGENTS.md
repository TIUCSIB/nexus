# AGENTS.md - Nexus 项目开发规范

## 项目概述

Nexus 是一个轻量级代理节点管理面板，采用 Go 后端 + Vue 3 前端 + sing-box 内核。面向个人/小圈子使用，无公开注册、无支付系统。

## 技术栈

- **后端**: Go 1.22+ / Gin / GORM / SQLite
- **前端**: Vue 3 / TypeScript / Pinia / shadcn-vue / Tailwind CSS
- **节点代理**: Go / sing-box
- **通信**: REST API（面板与Agent）

## 目录结构

`
Nexus/
  cmd/nexus/main.go              # 面板入口
  internal/
    config/                    # 配置加载
    database/                  # 数据库初始化和迁移
    model/                     # 数据模型（GORM）
    http/
      handler/               # HTTP 处理器
      middleware/            # JWT / Admin 中间件
      router/router.go       # 路由定义
    service/                   # 业务逻辑层
    subscription/              # 订阅格式生成（singbox/clash/universal）
    pkg/                       # 工具包（JWT、密码、加密）
  agent/                         # 节点代理（独立项目）
    cmd/agent/main.go          # Agent 入口
    install.sh                 # 一键安装脚本
    config.yml.example         # 示例配置
    internal/
      config/                # Agent 配置（支持 YAML/CLI/环境变量）
      httpclient/            # 面板通信客户端
      proxy/singbox.go       # sing-box 进程管理
      collector/stats.go     # 流量采集
      devicelimit/           # 设备限制执行器
  web/                           # 前端项目
    src/
      api/                   # API 调用模块
      views/                 # 页面组件
      components/ui/         # shadcn-vue 组件
      stores/                # Pinia 状态管理
      types/                 # TypeScript 类型定义
      utils/                 # 工具函数
      router/                # 路由配置
    vite.config.ts
  config.yaml                    # 面板配置文件
  Makefile
`

## 开发规范

### Go 后端

#### 代码风格
- 使用 gofmt 格式化代码
- 函数命名使用 CamelCase
- 接口名以 I 开头（如 IUserService）
- 错误处理必须显式检查，不要忽略错误返回值

#### 数据库
- 所有模型定义在 internal/model/ 目录
- 使用 GORM 的 AutoMigrate 自动迁移，不要手动写 SQL
- 新增字段必须设置合理的默认值
- 涉及金额的字段单位为分（int64）

#### API 设计
- 所有路由前缀为 /api/
- 认证接口: /api/auth/
- 用户接口: /api/user/（需 JWT）
- 管理接口: /api/admin/（需 JWT + Admin）
- Agent 接口: /api/internal/agent/（需 Token）
- 订阅接口: 动态路径，默认 /api/s/（需订阅 Token）

#### 响应格式
`go
// 成功
{"code": 0, "message": "success", "data": {...}}
// 错误
{"code": -1, "message": "错误信息"}
`

#### Handler 规范
- 请求体结构体定义在 handler 文件顶部，命名为 createXxxRequest / updateXxxRequest
- 使用 c.ShouldBindJSON 绑定请求体
- 使用统一的响应函数：Success(), BadRequest(), NotFound(), InternalError()
- 分页使用 parsePagination(c) 工具函数

#### 新增模型字段检查清单
1. 更新 internal/model/ 中的模型结构体
2. 更新 internal/http/handler/ 中对应的 create/update request 结构体
3. 更新 AdminCreateXxx 和 AdminUpdateXxx handler
4. 重启后端让 GORM AutoMigrate 生效

### Vue 前端

#### 文件编码
- 所有 .vue 和 .ts 文件必须使用 UTF-8 无 BOM 编码写入
- 使用 PowerShell 写入时必须用：
  [System.IO.File]::WriteAllText(, , [System.Text.UTF8Encoding](False))
- 绝对不要用 Out-File 或 Set-Content（会加 BOM 导致中文乱码）

#### 组件规范
- UI 组件使用 shadcn-vue
- 已安装的组件：button, card, input, label, badge, table, dialog, select, tabs, switch, textarea, dropdown-menu, sidebar, separator, collapsible, sonner, tooltip, skeleton, form, checkbox, pagination, sheet, avatar

#### API 调用
- API 模块放在 web/src/api/ 目录
- 使用 request 工具（Axios 封装），自动添加 JWT Header
- 所有 API 路径必须带 /api/ 前缀
- 响应格式统一为 ApiResponse<T> 类型

#### 页面开发
- 页面组件放在 web/src/views/ 对应子目录
- 路由配置在 web/src/router/index.ts
- 类型定义在 web/src/types/index.ts
- CRUD 操作必须有 try/catch 错误处理和 toast 提示
- 保存按钮要有 loading 状态防止重复提交

#### 中文文本
- 所有界面文本使用中文
- 不要使用 Unicode 转义序列
- 直接写中文字符

### 路由与侧边栏

#### 侧边栏结构
`
仪表盘
用户管理
套餐管理
节点管理（可折叠子菜单）
  节点管理
  权限组管理
  路由管理
系统设置
`

## 构建与运行

### 面板
`powershell
# 编译
go build -o bin/nexus.exe ./cmd/nexus/
# 运行
./bin/nexus.exe
`

### 前端
`powershell
cd web
npm run dev          # 开发服务器
npm run build        # 生产构建
`

### Agent
`powershell
cd agent
go build -o agent.exe ./cmd/agent/

# 方式一：命令行参数（推荐）
./agent.exe --panel https://panel.com --token REGISTER_TOKEN --name my-node

# 方式二：环境变量
="https://panel.com"
="REGISTER_TOKEN"
./agent.exe

# 方式三：配置文件
./agent.exe -config agent.yaml
`

#### Linux 一键安装
`ash
curl -fsSL https://your-panel.com/install-agent.sh | bash -s -- \
  --panel https://panel.com \
  --token REGISTER_TOKEN \
  --name my-node
`

## 数据库

- 数据库文件：data/nexus.db（SQLite）
- 使用 GORM AutoMigrate，启动时自动迁移
- 新增表或字段只需更新 model 结构体，重启后自动生效
- 不要手动修改数据库文件

## 测试账号

- 邮箱：admin@nexus.com
- 密码：12345678

## 常见问题

### 中文乱码
原因：PowerShell Out-File 会加 BOM 头
解决：始终用 [System.IO.File]::WriteAllText(, , [System.Text.UTF8Encoding](False))

### 前端路由冲突
原因：Vite 代理规则和前端路由冲突
解决：所有后端 API 路径必须带 /api/ 前缀

### 登录失败
原因：JWT token 过期或格式错误
解决：清除 localStorage 重新登录

### 后端重启后数据库报错
原因：新增字段没有默认值
解决：确保 model 字段有 gorm:"default:xxx" 标签

## 当前数据模型

### users 用户表
- id, uuid, email, password_hash, balance, plan_id, group_id
- traffic_used, traffic_limit, expired_at, is_admin, token, status
- device_limit, speed_limit_up, speed_limit_down

### plans 套餐表
- id, name, description, group_id, traffic_limit, duration_days, price
- speed_limit, device_limit, capacity_limit, traffic_reset
- sort, status

### nodes 节点表
- id, name, address, protocol, port, group_id, route_id
- rate, dynamic_rate, tags, traffic_limit, traffic_used, online_count
- parent_id, security, transport, flow_control, vless_encryption
- config_mode, config_json, online, last_heartbeat, register_token
- sort, status

### server_groups 权限组表
- id, name

### route_rules 路由规则表
- id, name, match, action, action_value, sort, status

### traffic_logs 流量记录表
- id, user_id, node_id, upload, download, recorded_at

### alive_ips 在线IP表
- id, user_id, ips, node_id, updated_at

### system_configs 系统配置表（KV）
- key, value

### node_auths 节点认证表
- id, node_id, auth_token

## 核心功能说明

### 订阅路径
- 支持自定义订阅路径，默认为 /s/
- 用户订阅 URL 格式：https://域名/{sub_path}/{token}
- 路径可在系统设置中修改，修改后立即生效（无需重启）
- 同时支持 API 格式：/api/{sub_path}/{token}?format=singbox

### 设备限制
- Agent 每 10 秒检查一次在线连接
- 超出设备限制的用户，多余连接会被自动关闭
- Agent 每 60 秒从面板同步最新的设备限制配置

### 流量重置
- 方式 0：不重置
- 方式 1：每月 1 号重置
- 方式 2：按用户周期重置（基于开通时间）
- 方式 3：每年 1 月 1 日重置

### 订阅信息节点
- 开启后会在订阅节点列表末尾添加两个信息节点
- 显示套餐到期时间和剩余流量

### 强制 HTTPS
- 开启后所有 HTTP 请求自动跳转到 HTTPS
