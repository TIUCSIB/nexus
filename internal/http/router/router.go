package router

import (
	"strings"
	"time"

	"nexus/internal/database"
	"nexus/internal/http/handler"
	"nexus/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()

	// 从设置中读取路径前缀（接口伪装，防止被识别为代理面板）
	adminPath := database.GetSettingDefault("admin_path", "admin")
	authPath := database.GetSettingDefault("auth_path", "auth")
	userPath := database.GetSettingDefault("user_path", "user")

	// Global middleware chain
	r.Use(middleware.ForceHTTPS())
	r.Use(middleware.SecurityHeaders())

	// Public site info (no auth required, used by login page)
	r.GET("/api/site/info", handler.GetSiteInfo)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "nexus"})
	})

	// Auth (no registration - admin creates users)
	auth := r.Group("/api/" + authPath)
	auth.Use(middleware.RateLimit(5, time.Minute))
	{
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.RefreshToken)
	}

	// User endpoints (JWT required)
	user := r.Group("/api/" + userPath)
	user.Use(middleware.Auth(), middleware.RateLimit(30, time.Minute))
	{
		user.GET("/profile", handler.GetProfile)
		user.PUT("/profile", handler.UpdateProfile)
		user.GET("/subscription", handler.GetSubscription)
		user.GET("/stats", handler.GetUserStats)
	}

	// Node endpoints (JWT required)
	nodes := r.Group("/api/nodes")
	nodes.Use(middleware.Auth(), middleware.RateLimit(30, time.Minute))
	{
		nodes.GET("", handler.ListNodes)
		nodes.GET("/:id/status", handler.NodeStatus)
	}

	// Agent communication (internal, server_token-based)
	agent := r.Group("/api/internal/agent")
	agent.Use(handler.ServerAuthMiddleware())
	{
		agent.POST("/handshake", handler.AgentHandshake)
		agent.GET("/:node_id/config", handler.AgentGetConfig)
		agent.GET("/:node_id/users", handler.AgentGetUsers)
		agent.POST("/:node_id/heartbeat", handler.AgentHeartbeat)
		agent.POST("/:node_id/report", handler.AgentReport)
		agent.POST("/:node_id/traffic", handler.AgentReportTraffic)
		agent.POST("/:node_id/alive", handler.AgentReportAlive)
		agent.GET("/alivelist", handler.AgentGetAliveList)
		agent.GET("/devicelimit", handler.AgentGetDeviceLimit)
	}

	// WebSocket endpoint (self-authenticated, outside middleware)
	r.GET("/api/internal/agent/ws", handler.AgentWebSocket)

	// Machine endpoints (internal, machine_token-based)
	machine := r.Group("/api/internal/machine")
	machine.Use(handler.MachineAuthMiddleware())
	{
		machine.POST("/:id/heartbeat", handler.MachineHeartbeat)
		machine.GET("/:id/nodes", handler.MachineGetNodes)
		machine.POST("/:id/load", handler.MachineReportLoad)
	}

	// Admin endpoints (JWT + admin required)
	admin := r.Group("/api/" + adminPath)
	admin.Use(middleware.Auth(), middleware.Admin(), middleware.RateLimit(60, time.Minute))
	{
		admin.GET("/users", handler.AdminListUsers)
		admin.GET("/users/:id", handler.AdminGetUser)
		admin.POST("/users", handler.AdminCreateUser)
		admin.PUT("/users/:id", handler.AdminUpdateUser)
		admin.DELETE("/users/:id", handler.AdminDeleteUser)
		admin.GET("/users/:id/traffic-logs", handler.AdminGetUserTrafficLogs)
		admin.GET("/users/:id/online-ips", handler.AdminGetUserOnlineIPs)
		admin.POST("/users/:id/reset-uuid", handler.AdminResetUserUUID)
		admin.POST("/users/:id/reset-traffic", handler.AdminResetUserTraffic)
		admin.POST("/users/batch-operation", handler.AdminBatchUserOperation)

		admin.GET("/traffic-reset/users", handler.AdminTrafficResetUsers)
		admin.POST("/traffic-reset/manual", handler.AdminManualTrafficReset)
		admin.GET("/traffic-reset/stats", handler.AdminTrafficResetStats)

		admin.GET("/audit-logs", handler.AdminListAuditLogs)

		admin.GET("/plans", handler.AdminListPlans)
		admin.POST("/plans", handler.AdminCreatePlan)
		admin.PUT("/plans/:id", handler.AdminUpdatePlan)
		admin.DELETE("/plans/:id", handler.AdminDeletePlan)

		admin.GET("/nodes", handler.AdminListNodes)
		admin.POST("/nodes", handler.AdminCreateNode)
		admin.POST("/nodes/generate-reality-keys", handler.AdminGenerateRealityKeys)
		admin.POST("/nodes/generate-ech-key", handler.AdminGenerateECHKey)
		admin.POST("/nodes/batch-delete", handler.AdminBatchDeleteNodes)
		admin.POST("/nodes/batch-reset-traffic", handler.AdminBatchResetNodeTraffic)
		admin.POST("/nodes/batch-update", handler.AdminBatchUpdateNodes)
		admin.POST("/nodes/sort", handler.AdminSortNodes)
		admin.POST("/nodes/copy", handler.AdminCopyNode)
		admin.PUT("/nodes/:id", handler.AdminUpdateNode)
		admin.DELETE("/nodes/:id", handler.AdminDeleteNode)
		admin.POST("/nodes/:id/restart", handler.AdminRestartNode)
		admin.POST("/nodes/:id/reset-traffic", handler.AdminResetNodeTraffic)
		admin.POST("/nodes/:id/command", handler.AdminSendNodeCommand)
		admin.GET("/nodes/:id/outbounds", handler.AdminListNodeOutbounds)
		admin.PUT("/nodes/:id/outbounds", handler.AdminUpdateNodeOutbounds)

		admin.GET("/custom-outbounds", handler.AdminListCustomOutbounds)
		admin.POST("/custom-outbounds", handler.AdminCreateCustomOutbound)
		admin.PUT("/custom-outbounds/:id", handler.AdminUpdateCustomOutbound)
		admin.DELETE("/custom-outbounds/:id", handler.AdminDeleteCustomOutbound)

		admin.GET("/machines", handler.AdminListMachines)
		admin.GET("/machines/:id", handler.AdminGetMachine)
		admin.GET("/machines/:id/history", handler.AdminGetMachineLoadHistory)
		admin.POST("/machines", handler.AdminCreateMachine)
		admin.PUT("/machines/:id", handler.AdminUpdateMachine)
		admin.DELETE("/machines/:id", handler.AdminDeleteMachine)
		admin.POST("/machines/:id/reset-token", handler.AdminResetMachineToken)
		admin.POST("/machines/:id/install-command", handler.AdminGetMachineInstallCommand)
		admin.GET("/machines/:id/nodes", handler.AdminListMachineNodes)

		admin.GET("/groups", handler.AdminListGroups)
		admin.POST("/groups", handler.AdminCreateGroup)
		admin.PUT("/groups/:id", handler.AdminUpdateGroup)
		admin.DELETE("/groups/:id", handler.AdminDeleteGroup)

		admin.GET("/routes", handler.AdminListRoutes)
		admin.POST("/routes", handler.AdminCreateRoute)
		admin.PUT("/routes/:id", handler.AdminUpdateRoute)
		admin.DELETE("/routes/:id", handler.AdminDeleteRoute)

		admin.GET("/settings", handler.AdminGetSettings)
		admin.GET("/settings/subscription-template-defaults", handler.AdminGetSubscriptionTemplateDefaults)
		admin.PUT("/settings", handler.AdminUpdateSettings)
		admin.POST("/settings/backup", handler.AdminBackupDatabase)
		admin.GET("/settings/backup-info", handler.AdminBackupInfo)

		admin.GET("/stats/overview", handler.AdminStatsOverview)
		admin.GET("/stats/traffic", handler.AdminStatsTraffic)
		admin.GET("/stats/node-ranking", handler.AdminNodeTrafficRanking)
		admin.GET("/stats/user-ranking", handler.AdminUserTrafficRanking)
		admin.GET("/stats/system", handler.AdminSystemStatus)

		admin.GET("/users/:id/usage", handler.AdminUserUsageDetails)

		admin.GET("/online-ips", handler.AdminListOnlineIPs)
		admin.GET("/traffic-logs", handler.AdminListTrafficLogs)
	}

	// Serve static assets (JS, CSS, etc.)
	r.Static("/assets", "./web/dist/assets")

	// SPA catch-all + subscription routing
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}

		path := strings.Trim(c.Request.URL.Path, "/")
		if strings.HasPrefix(path, "api/") {
			handler.SubRouter(c)
			return
		}

		// Check subscription path: /{sub_path}/{token}
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 {
			subPath := strings.Trim(database.GetSettingDefault("sub_path", "s"), "/")
			if subPath == "" {
				subPath = "s"
			}
			if parts[0] == subPath {
				q := c.Request.URL.Query()
				q.Set("token", parts[1])
				c.Request.URL.RawQuery = q.Encode()
				handler.SubAutoDetect(c)
				return
			}
		}

		// Fallback to SPA
		c.File("./web/dist/index.html")
	})

	return r
}
