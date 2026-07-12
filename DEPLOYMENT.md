# Nexus 多环境部署指南

本项目支持**两种部署方式共存**，代码无需修改，只需选择不同的构建/部署命令。

---

## 方案对比

| | **方案 1：VPS 全栈部署** | **方案 2：Serverless 混合部署** |
|--|------------------------|------------------------------|
| **前端** | VPS 上（内嵌或 Nginx） | Cloudflare Pages |
| **后端** | VPS 上（Go 进程） | Fly.io |
| **数据库** | VPS 上 SQLite | Fly.io 持久卷 SQLite |
| **节点** | 独立 VPS | 独立 VPS（两方案可共享） |
| **费用** | $5~10/月 | 前端免费 + 后端免费（<100GB） |
| **适合** | 个人、稳定优先 | 全球快、高可用 |

---

## 方案 1：VPS 全栈部署

### 架构
```
VPS A (面板)
  ├── nexus (Go 后端 + 内嵌前端)
  ├── web/dist/
  └── data/nexus.db

VPS B/C/... (节点)
  └── nexus-agent + sing-box
```

### 部署步骤

#### 1. 构建前端（VPS 模式）
```bash
cd web
npm install
npm run build:vps
# 输出到 web/dist/
```

**或直接用默认 build：**
```bash
# 修改 .env.production.vps 里的 IP
npm run build
```

#### 2. 上传到服务器
```bash
# 压缩
cd ..
tar czf nexus-vps.tar.gz cmd/ internal/ pkg/ go.mod go.sum config.yaml web/dist/

# 上传
scp nexus-vps.tar.gz root@你的VPS:/opt/

# 或用 GitHub Release 下载
```

#### 3. 服务器上编译运行
```bash
ssh root@你的VPS
cd /opt
tar xzf nexus-vps.tar.gz
mv nexus-vps nexus && cd nexus

# 编译（或下载预编译）
go build -o nexus ./cmd/nexus/
# 或
wget https://github.com/TIUCSIB/nexus/releases/latest/download/nexus-linux-amd64
chmod +x nexus-linux-amd64 && mv nexus-linux-amd64 nexus

# 配置
cat > config.yaml <<EOF
server:
  port: 20241
  host: 0.0.0.0
database:
  path: data/nexus.db
jwt:
  secret: $(openssl rand -hex 32)
web:
  static_dir: web/dist
EOF

# 启动
nohup ./nexus -config config.yaml >> nexus.log 2>&1 &
```

#### 4. 节点配置
```bash
# 节点机 /opt/nexus-agent/agent.yaml
panel:
  url: http://你的面板VPS_IP:20241
  token: 面板里生成的注册Token
```

---

## 方案 2：Serverless 混合部署

### 架构
```
CF Pages (前端)
  ↓
Fly.io (后端)
  ↓
节点 VPS (Agent + sing-box)
```

### 部署步骤

#### 1. 前端 → Cloudflare Pages

```bash
cd web

# 构建（Serverless 模式）
npm run build:serverless
# 输出到 web/dist/

# 方式 1：Wrangler CLI
npm install -g wrangler
wrangler login
wrangler pages deploy dist --project-name=nexus-panel

# 方式 2：GitHub 集成（推荐）
# 1. 推代码到 GitHub
# 2. CF Dashboard → Pages → Connect to Git → TIUCSIB/nexus
# 3. 构建设置：
#    - Build command: cd web && npm install && npm run build:serverless
#    - Build output: web/dist
#    - 环境变量: VITE_API_BASE_URL=https://nexus-api.fly.dev
```

访问：`https://nexus-panel.pages.dev`

---

#### 2. 后端 → Fly.io

```bash
cd /c/Users/biscuit/Documents/Nexus

# 安装 Fly CLI
# Windows: https://fly.io/docs/flyctl/install/
# macOS: brew install flyctl
# Linux: curl -L https://fly.io/install.sh | sh

# 登录
fly auth login

# 初始化（会创建 fly.toml）
fly launch --no-deploy
# 选择应用名：nexus-api
# 选择区域：sin (Singapore) 或离你近的
# 不要现在部署：No

# 创建持久卷（存 SQLite）
fly volumes create nexus_data --size 1

# 设置密钥
fly secrets set JWT_SECRET=$(openssl rand -hex 32)

# 构建前端（Serverless 模式）
cd web && npm run build:serverless && cd ..

# 修改 Dockerfile（包含前端）
cat > Dockerfile <<'EOF'
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o nexus ./cmd/nexus/

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata sqlite
WORKDIR /app
COPY --from=builder /app/nexus .
COPY web/dist ./web/dist
RUN mkdir -p /data
EXPOSE 8080
CMD ["./nexus", "-config", "/app/config.yaml"]
EOF

# 确保 config.yaml 使用环境变量和正确路径
cat > config.yaml <<'EOF'
server:
  port: 8080
  host: 0.0.0.0
database:
  path: /data/nexus.db
jwt:
  secret: ${JWT_SECRET}
web:
  static_dir: web/dist
EOF

# 部署
fly deploy

# 查看状态
fly status
fly logs

# 获取访问地址
fly open
# 例如：https://nexus-api.fly.dev
```

---

#### 3. 更新前端环境变量

回到 `.env.production.serverless`，改成实际 Fly.io 地址：

```bash
# web/.env.production.serverless
VITE_API_BASE_URL=https://nexus-api.fly.dev
```

重新构建并部署到 CF Pages。

---

#### 4. 节点配置

```bash
# 节点机 /opt/nexus-agent/agent.yaml
panel:
  url: https://nexus-api.fly.dev
  token: 面板里生成的注册Token
```

---

## 在两种方案间切换

### 从 VPS 迁移到 Serverless

1. **备份数据库**：
   ```bash
   # VPS 上
   scp root@你的VPS:/opt/nexus/data/nexus.db ./nexus.db.backup
   ```

2. **导入到 Fly.io**：
   ```bash
   # 临时挂载卷
   fly ssh console
   # 在 Fly.io shell 里：
   cd /data
   # 退出，从本地上传
   fly ssh sftp shell
   put nexus.db.backup /data/nexus.db
   exit
   ```

3. **重启 Fly.io**：
   ```bash
   fly apps restart nexus-api
   ```

4. **更新节点 Agent 配置**（改 `panel.url`）

---

### 从 Serverless 迁移回 VPS

1. **从 Fly.io 下载数据库**：
   ```bash
   fly ssh sftp shell
   get /data/nexus.db nexus.db
   exit
   ```

2. **上传到 VPS**：
   ```bash
   scp nexus.db root@你的VPS:/opt/nexus/data/
   ```

3. **更新节点 Agent 配置**（改回 VPS IP）

---

## 日常更新

### 更新代码后

**VPS 方案：**
```bash
# 本地
cd web && npm run build:vps && cd ..
# 上传 web/dist + 重新编译后端
```

**Serverless 方案：**
```bash
# 前端
cd web && npm run build:serverless
wrangler pages deploy dist --project-name=nexus-panel

# 后端
cd .. && fly deploy
```

---

## 常见问题

### Q: 两套方案可以同时运行吗？
**A:** 可以。前端可以部署两份（一份在 VPS，一份在 CF Pages），节点可以同时连两个面板（不同 token）。

### Q: 节点能在两套方案间共享吗？
**A:** 可以。只需修改 `agent.yaml` 里的 `panel.url`，指向 VPS 或 Fly.io。

### Q: Fly.io 睡眠怎么办？
**A:** 免费版会自动休眠，首次访问唤醒（3~5秒）。升级到 Hobby ($5/月) 永不休眠。

### Q: 数据库能共享吗？
**A:** 不能。VPS 和 Fly.io 各有独立的 SQLite 文件，需要手动同步（见迁移步骤）。

---

## 推荐方案选择

| 你的情况 | 推荐 |
|---------|------|
| 个人使用，追求简单 | **方案 1（VPS）** |
| 想体验 serverless | **方案 2** |
| 多地用户，要快 | **方案 2（前端 CF CDN）** |
| 完全免费 | **方案 2（Fly.io 免费额度）** |
| 稳定第一 | **方案 1** |

---

## 技术细节

### 前端环境变量优先级
```
.env.production.vps         # npm run build:vps
.env.production.serverless  # npm run build:serverless
.env.production             # npm run build (默认)
.env.development            # npm run dev
```

### 后端配置适配
两种方案用同一份 `config.yaml`，区别只在：
- VPS: `database.path` 相对路径 `data/nexus.db`
- Fly.io: `database.path` 绝对路径 `/data/nexus.db`（挂载卷）

### 节点无缝切换
Agent 只关心 `panel.url`，协议相同（HTTP/HTTPS + REST API），所以：
- VPS 面板: `http://IP:20241`
- Fly.io 面板: `https://nexus-api.fly.dev`

两者 API 完全兼容，改配置即可切换。

---

## 总结

✅ **同一套代码**  
✅ **两种部署方式**  
✅ **随时切换**  
✅ **节点可共享**  

选择你喜欢的方式，或两个都试试！
