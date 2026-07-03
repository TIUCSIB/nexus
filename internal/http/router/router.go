package router

import (
	"strings"

	"nexus/internal/database"
	"nexus/internal/http/handler"
	"nexus/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()

	// Force HTTPS
	r.Use(middleware.ForceHTTPS())

	// Public site info (no auth required, used by login page)
	r.GET("/api/site/info", handler.GetSiteInfo)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "nexus"})
	})

	// Auth (no registration - admin creates users)
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.RefreshToken)
	}

	// User endpoints (JWT required)
	user := r.Group("/api/user")
	user.Use(middleware.Auth())
	{
		user.GET("/profile", handler.GetProfile)
		user.PUT("/profile", handler.UpdateProfile)
		user.GET("/subscription", handler.GetSubscription)
	}

	// Node endpoints (JWT required)
	nodes := r.Group("/api/nodes")
	nodes.Use(middleware.Auth())
	{
		nodes.GET("", handler.ListNodes)
		nodes.GET("/:id/status", handler.NodeStatus)
	}

	// Agent communication (internal, token-based)
	agent := r.Group("/api/internal/agent")
	{
		agent.POST("/register", handler.AgentRegister)
		agentAuth := agent.Group("")
		agentAuth.Use(handler.AgentAuthMiddleware())
		{
			agentAuth.POST("/heartbeat", handler.AgentHeartbeat)
			agentAuth.GET("/config", handler.AgentGetConfig)
			agentAuth.POST("/traffic", handler.AgentReportTraffic)
			agentAuth.POST("/alive", handler.AgentReportAlive)
			agentAuth.GET("/alivelist", handler.AgentGetAliveList)
			agentAuth.GET("/devicelimit", handler.AgentGetDeviceLimit)
		}
	}

	// Admin endpoints (JWT + admin required)
	admin := r.Group("/api/admin")
	admin.Use(middleware.Auth(), middleware.Admin())
	{
		admin.GET("/users", handler.AdminListUsers)
		admin.GET("/users/:id", handler.AdminGetUser)
		admin.POST("/users", handler.AdminCreateUser)
		admin.PUT("/users/:id", handler.AdminUpdateUser)
		admin.DELETE("/users/:id", handler.AdminDeleteUser)
		admin.GET("/users/:id/traffic-logs", handler.AdminGetUserTrafficLogs)

		admin.GET("/plans", handler.AdminListPlans)
		admin.POST("/plans", handler.AdminCreatePlan)
		admin.PUT("/plans/:id", handler.AdminUpdatePlan)
		admin.DELETE("/plans/:id", handler.AdminDeletePlan)

		admin.GET("/nodes", handler.AdminListNodes)
		admin.POST("/nodes", handler.AdminCreateNode)
		admin.POST("/nodes/generate-reality-keys", handler.AdminGenerateRealityKeys)
		admin.PUT("/nodes/:id", handler.AdminUpdateNode)
		admin.DELETE("/nodes/:id", handler.AdminDeleteNode)
		admin.POST("/nodes/:id/restart", handler.AdminRestartNode)
		admin.POST("/nodes/:id/reset-traffic", handler.AdminResetNodeTraffic)

		admin.GET("/groups", handler.AdminListGroups)
		admin.POST("/groups", handler.AdminCreateGroup)
		admin.PUT("/groups/:id", handler.AdminUpdateGroup)
		admin.DELETE("/groups/:id", handler.AdminDeleteGroup)

		admin.GET("/routes", handler.AdminListRoutes)
		admin.POST("/routes", handler.AdminCreateRoute)
		admin.PUT("/routes/:id", handler.AdminUpdateRoute)
		admin.DELETE("/routes/:id", handler.AdminDeleteRoute)

		admin.GET("/settings", handler.AdminGetSettings)
		admin.PUT("/settings", handler.AdminUpdateSettings)

		admin.GET("/stats/overview", handler.AdminStatsOverview)
		admin.GET("/stats/traffic", handler.AdminStatsTraffic)

		admin.GET("/online-ips", handler.AdminListOnlineIPs)
		admin.GET("/traffic-logs", handler.AdminListTrafficLogs)
	}


	// Serve static assets (JS, CSS, etc.)
	r.Static("/assets", "./web/dist/assets")

	// SPA catch-all闁挎稒鑹鹃幃鎾诲籍鐠轰警妲遍柣鐐叉閻楀鎹勯姘辩獮妤犵偞褰冮崳锝夊冀閻撳海纭€闁汇劌瀚褰掓⒓閸涘瓨鎳犻柟?	// 濞撴艾顑呴々褔鏁?{sub_path}/{token} -> /s/f034db92-8d33-4952-9cd5-2fe01669a379
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}

		// 閻忓繑绻嗛惁顖炲礌瑜版帒甯抽柡宥囶攰閻儳顕ラ崟顒傚鐎殿喖楠忕槐?{sub_path}/{token}
		path := strings.Trim(c.Request.URL.Path, "/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 {
			subPath := strings.Trim(database.GetSettingDefault("sub_path", "s"), "/")
			if subPath == "" {
				subPath = "s"
			}
			if parts[0] == subPath {
				// 闁告牕缍婇崢銈夊礆閹峰矈鍚傞梻鍐ㄦ嚀閻儳顕ラ崟鍓佺閻?token 闁衡偓閹冨汲 query 妤犵偞鍎肩换鎴﹀炊?singbox 闁哄秶鍘х槐?				c.Request.URL.RawQuery = "token=" + parts[1]
				handler.SubSingbox(c)
				return
			}
		}

		// 闁稿繑婀圭划顒傛崉椤栨氨绐為悹?SPA
		c.File("./web/dist/index.html")
	})

	return r
}