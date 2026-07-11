# GitHub Actions 工作流说明

本项目使用 GitHub Actions 自动构建和发布。

## 工作流

### 1. Build and Test (build.yml)
**触发条件：**
- Push 到 `main` 或 `dev` 分支
- 创建 Pull Request 到 `main` 或 `dev` 分支

**功能：**
- 运行后端测试
- 运行前端构建测试
- 构建 Docker 镜像测试

### 2. Build and Release (release.yml)
**触发条件：**
- 推送标签（格式：`v*`，例如 `v1.0.0`）
- 手动触发（通过 GitHub Actions 界面）

**构建产物：**
- **面板二进制文件**
  - `nexus-linux-amd64`
  - `nexus-linux-arm64`
  - `nexus-windows-amd64.exe`
  - `nexus-darwin-amd64`
  - `nexus-darwin-arm64`

- **Agent 二进制文件**
  - `nexus-agent-linux-amd64`
  - `nexus-agent-linux-arm64`
  - `nexus-agent-windows-amd64.exe`
  - `nexus-agent-darwin-amd64`
  - `nexus-agent-darwin-arm64`

- **NS CLI 工具**
  - `ns-linux-amd64`
  - `ns-linux-arm64`
  - `ns-windows-amd64.exe`
  - `ns-darwin-amd64`
  - `ns-darwin-arm64`

- **前端静态文件**
  - `web-dist.zip`
  - `web-dist.tar.gz`

所有文件会自动发布到 GitHub Releases。

## 如何发布新版本

### 方式一：使用标签（推荐）

```bash
# 1. 确保所有修改已提交
git add .
git commit -m "feat: 添加新功能"
git push

# 2. 创建并推送标签
git tag v1.0.0
git push origin v1.0.0

# 3. GitHub Actions 会自动构建并创建 Release
```

### 方式二：手动触发

1. 访问 GitHub 仓库页面
2. 点击 **Actions** 标签
3. 选择 **Build and Release** 工作流
4. 点击 **Run workflow** 按钮
5. 选择分支并运行

## 版本号规范

推荐使用语义化版本号：

- `v1.0.0` - 主版本（重大更新）
- `v1.1.0` - 次版本（新增功能）
- `v1.0.1` - 补丁版本（修复 bug）

## 安装脚本自动更新

安装脚本会自动从 GitHub Releases 下载最新版本：

```bash
# 安装面板（自动获取最新版本）
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/install-panel.sh)

# 安装指定版本
bash <(curl -fsSL https://raw.githubusercontent.com/TIUCSIB/nexus-install/master/install-panel.sh) v1.0.0
```

## 本地构建

如果需要本地构建：

```bash
# 构建面板
go build -o nexus ./cmd/nexus

# 构建 Agent
cd agent
go build -o nexus-agent ./cmd/agent
go build -o ns ./cmd/ns

# 构建前端
cd web
npm install
npm run build
```

## 构建 Docker 镜像

```bash
# 构建
docker build -t nexus:latest .

# 或使用 docker-compose
docker-compose build
```

## 故障排查

### 构建失败
1. 检查 Go 版本是否为 1.22+
2. 检查 Node.js 版本是否为 20+
3. 查看 Actions 日志获取详细错误信息

### CGO 错误
- Linux ARM64 构建需要交叉编译工具链
- 工作流已配置 `gcc-aarch64-linux-gnu`

### 前端构建错误
- 检查 `web/package-lock.json` 是否存在
- 确保依赖项已正确安装

## 相关链接

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [语义化版本规范](https://semver.org/lang/zh-CN/)
