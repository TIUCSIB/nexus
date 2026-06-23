package router

import (
	"nexus/internal/http/handler"
	"nexus/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "nexus"})
	})

	v1 := r.Group("/api/v1")
	{
		// Auth (no registration - admin creates users)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handler.Login)
			auth.POST("/refresh", handler.RefreshToken)
		}

		// Subscription (token-based, no JWT)
		sub := v1.Group("/sub")
		{
			sub.GET("/singbox", handler.SubSingbox)
			sub.GET("/clash", handler.SubClash)
			sub.GET("/surge", handler.SubSurge)
			sub.GET("/surfboard", handler.SubSurfboard)
			sub.GET("/shadowrocket", handler.SubShadowrocket)
			sub.GET("/v2rayn", handler.SubV2RayN)
		}

		// User endpoints (JWT required)
		user := v1.Group("/user")
		user.Use(middleware.Auth())
		{
			user.GET("/profile", handler.GetProfile)
			user.PUT("/profile", handler.UpdateProfile)
			user.GET("/subscription", handler.GetSubscription)
		}

		// Node endpoints (JWT required)
		nodes := v1.Group("/nodes")
		nodes.Use(middleware.Auth())
		{
			nodes.GET("", handler.ListNodes)
			nodes.GET("/:id/status", handler.NodeStatus)
		}

		// Agent communication (internal, token-based)
		agent := v1.Group("/internal/agent")
		{
			// Register uses the register token directly (no auth middleware)
			agent.POST("/register", handler.AgentRegister)

			// All other agent endpoints require a valid node auth token
			agentAuth := agent.Group("")
			agentAuth.Use(handler.AgentAuthMiddleware())
			{
				agentAuth.POST("/heartbeat", handler.AgentHeartbeat)
				agentAuth.GET("/config", handler.AgentGetConfig)
				agentAuth.POST("/traffic", handler.AgentReportTraffic)
				agentAuth.POST("/alive", handler.AgentReportAlive)
				agentAuth.GET("/alivelist", handler.AgentGetAliveList)
			}
		}

		// Admin endpoints (JWT + admin required)
		admin := v1.Group("/admin")
		admin.Use(middleware.Auth(), middleware.Admin())
		{
			// User management
			admin.GET("/users", handler.AdminListUsers)
			admin.GET("/users/:id", handler.AdminGetUser)
			admin.POST("/users", handler.AdminCreateUser)
			admin.PUT("/users/:id", handler.AdminUpdateUser)
			admin.DELETE("/users/:id", handler.AdminDeleteUser)

			// Plan management
			admin.GET("/plans", handler.AdminListPlans)
			admin.POST("/plans", handler.AdminCreatePlan)
			admin.PUT("/plans/:id", handler.AdminUpdatePlan)
			admin.DELETE("/plans/:id", handler.AdminDeletePlan)

			// Node management
			admin.GET("/nodes", handler.AdminListNodes)
			admin.POST("/nodes", handler.AdminCreateNode)
			admin.PUT("/nodes/:id", handler.AdminUpdateNode)
			admin.DELETE("/nodes/:id", handler.AdminDeleteNode)
			admin.POST("/nodes/:id/restart", handler.AdminRestartNode)

			// Route rules management
			admin.GET("/routes", handler.AdminListRoutes)
			admin.POST("/routes", handler.AdminCreateRoute)
			admin.PUT("/routes/:id", handler.AdminUpdateRoute)
			admin.DELETE("/routes/:id", handler.AdminDeleteRoute)

			// Settings
			admin.GET("/settings", handler.AdminGetSettings)
			admin.PUT("/settings", handler.AdminUpdateSettings)

			// Stats
			admin.GET("/stats/overview", handler.AdminStatsOverview)
			admin.GET("/stats/traffic", handler.AdminStatsTraffic)
		}
	}

	return r
}
