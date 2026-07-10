# Agent 三大功能实现计划

## 功能 1：内核级限速（Hysteria2）

### 原理
sing-box 的 Hysteria2 inbound 支持 per-user `up_mbps`/`down_mbps` 字段。当前 `kernel.User` 携带了 `SpeedLimit` 但从未使用。

### 修改

**`agent/internal/kernel/config.go`**
- `hysteria2User` 结构体增加 `UpMbps int` 和 `DownMbps int` 字段（带 `json:"up_mbps,omitempty"` 等标签）
- `buildHysteria2Inbound` 中，将 `u.SpeedLimit` 赋值给用户的 `UpMbps` 和 `DownMbps`

**限制：** VLESS 和 TUIC 协议在 sing-box 中无 per-user 限速支持，跳过。

---

## 功能 2：WS 推送配置（跳过 HTTP 轮询）

### 架构

当 WS 连接正常时，面板主动推送配置/用户数据，agent 不再依赖 heartbeat 检测 configChanged。

```
面板更新节点配置
  → hub.SendCommand(nodeID, {type:"sync.config", data:{...}})
  → agent WS handler 解析数据
  → 发送到 wsUpdateCh channel
  → main loop 收到 → hotReloadOrRestart
```

### 面板端修改

**`internal/ws/hub.go`** — 新增两个方法：
```go
func (h *Hub) PushConfig(nodeID string, config interface{}) error
func (h *Hub) PushUsers(nodeID string, users interface{}) error
```
内部调用现有的 `SendCommand`，type 分别为 `"sync.config"` 和 `"sync.users"`。

**`internal/http/handler/agent.go`** — 在 `AgentGetConfig` 和 `AgentGetUsers` 的 ETag 缓存逻辑之外，当配置被管理员修改时触发推送。

**`internal/http/handler/admin_node.go`** — 在 `AdminUpdateNode` 成功后，调用 `WSHub.PushConfig(nodeID, config)` 推送新配置。

**`internal/http/handler/admin_user.go`** — 在用户 CRUD 操作后，向受影响的节点推送用户列表。

### Agent 端修改

**`agent/internal/wsclient/client.go`** — 新增：
- `connected atomic.Bool` 字段
- `IsConnected() bool` 方法
- 在 `Connect()` 成功后设置 `true`，在 `readPump` 退出时设置 `false`

**`agent/cmd/agent/main.go`** — 在 `runNode` 中：

1. 新增 channel：
```go
type wsUpdate struct {
    config *kernel.NodeConfigFromPanel
    users  []kernel.User
}
wsUpdateCh := make(chan wsUpdate, 1)
```

2. 注册新 WS handler：
- `"sync.config"` → 解析 config 数据 → 发送到 wsUpdateCh
- `"sync.users"` → 解析 users 数据 → 发送到 wsUpdateCh

3. main loop 新增 select case：
```go
case update := <-wsUpdateCh:
    // 缓存 config/users
    // 调用 hotReloadOrRestart
```

4. heartbeat case 中：当 `wsClient.IsConnected()` 为 true 时，跳过 `configChanged` 检测（只上报 stats），因为配置变更已通过 WS 推送。

---

## 功能 3：增量用户同步

### 简化方案

不实现完整的 delta tracking，而是：
- 面板在用户变更时通过 WS 推送**完整用户列表**（`sync.users`）
- Agent 收到后直接热重载

这比 delta 更简单可靠，且利用了功能 2 的 WS 通道。后续可优化为 delta。

### 面板端触发点

在以下操作后推送用户列表：
- `AdminCreateUser` / `AdminUpdateUser` / `AdminDeleteUser`
- `AdminResetUserTraffic`
- 用户套餐到期自动禁用（scheduler）

需要确定受影响的节点（根据用户的 group_id 匹配节点的 group_id），只向相关节点推送。

---

## 实现顺序

1. **kernel/config.go** — Hysteria2 per-user 限速（最简单，纯 agent 端）
2. **wsclient/client.go** — connected 状态追踪
3. **ws/hub.go** — PushConfig / PushUsers 方法
4. **agent main.go** — WS handler + channel + main loop 集成
5. **admin handler** — 在 CRUD 后触发 WS 推送
