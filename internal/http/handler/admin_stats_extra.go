package handler

import (
	"fmt"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminNodeTrafficRanking returns traffic usage for all nodes, ordered by total traffic.
// GET /api/admin/stats/node-ranking
func AdminNodeTrafficRanking(c *gin.Context) {
	var results []struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		Address      string `json:"address"`
		Protocol     string `json:"protocol"`
		TrafficUsed  int64  `json:"traffic_used"`
		TrafficLimit int64  `json:"traffic_limit"`
		Online       bool   `json:"online"`
		OnlineCount  int    `json:"online_count"`
	}

	database.DB.Model(&model.Node{}).
		Select("id, name, address, protocol, traffic_used, traffic_limit, online, online_count").
		Order("traffic_used DESC").
		Limit(50).
		Scan(&results)

	if results == nil {
		results = []struct {
			ID           uint   `json:"id"`
			Name         string `json:"name"`
			Address      string `json:"address"`
			Protocol     string `json:"protocol"`
			TrafficUsed  int64  `json:"traffic_used"`
			TrafficLimit int64  `json:"traffic_limit"`
			Online       bool   `json:"online"`
			OnlineCount  int    `json:"online_count"`
		}{}
	}

	Success(c, gin.H{
		"nodes": results,
	})
}

// AdminUserTrafficRanking returns top N users by traffic usage.
// GET /api/admin/stats/user-ranking?limit=20
func AdminUserTrafficRanking(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var results []struct {
		ID           uint   `json:"id"`
		Email        string `json:"email"`
		UUID         string `json:"uuid"`
		TrafficUsed  int64  `json:"traffic_used"`
		TrafficLimit int64  `json:"traffic_limit"`
		UploadUsed   int64  `json:"upload_used"`
		DownloadUsed int64  `json:"download_used"`
		DeviceLimit  int    `json:"device_limit"`
	}

	database.DB.Model(&model.User{}).
		Select("id, email, uuid, traffic_used, traffic_limit, upload_used, download_used, device_limit").
		Where("status = 1").
		Order("traffic_used DESC").
		Limit(limit).
		Scan(&results)

	if results == nil {
		results = []struct {
			ID           uint   `json:"id"`
			Email        string `json:"email"`
			UUID         string `json:"uuid"`
			TrafficUsed  int64  `json:"traffic_used"`
			TrafficLimit int64  `json:"traffic_limit"`
			UploadUsed   int64  `json:"upload_used"`
			DownloadUsed int64  `json:"download_used"`
			DeviceLimit  int    `json:"device_limit"`
		}{}
	}

	Success(c, gin.H{
		"users": results,
	})
}

// AdminUserUsageDetails returns detailed usage data for a specific user.
// GET /api/admin/users/:id/usage
func AdminUserUsageDetails(c *gin.Context) {
	userID := c.Param("id")

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		NotFound(c, "用户不存在")
		return
	}

	// Get per-node traffic breakdown
	var nodeTraffic []struct {
		NodeID   uint   `json:"node_id"`
		NodeName string `json:"node_name"`
		Upload   int64  `json:"upload"`
		Download int64  `json:"download"`
	}
	database.DB.Table("traffic_logs t").
		Select("t.node_id, n.name as node_name, SUM(t.upload) as upload, SUM(t.download) as download").
		Joins("LEFT JOIN nodes n ON n.id = t.node_id").
		Where("t.user_id = ?", user.ID).
		Group("t.node_id").
		Scan(&nodeTraffic)

	// Get daily traffic for last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var dailyTraffic []struct {
		Date     string `json:"date"`
		Upload   int64  `json:"upload"`
		Download int64  `json:"download"`
	}
	database.DB.Table("traffic_logs").
		Select("DATE(recorded_at) as date, SUM(upload) as upload, SUM(download) as download").
		Where("user_id = ? AND recorded_at >= ?", user.ID, thirtyDaysAgo).
		Group("DATE(recorded_at)").
		Order("date ASC").
		Scan(&dailyTraffic)

	// Get online IP info
	var onlineIPs []struct {
		NodeName string `json:"node_name"`
		IPs      string `json:"ips"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	database.DB.Table("alive_ips a").
		Select("n.name as node_name, a.ips, a.updated_at").
		Joins("LEFT JOIN nodes n ON n.id = a.node_id").
		Where("a.user_id = ?", user.ID).
		Scan(&onlineIPs)

	if nodeTraffic == nil {
		nodeTraffic = nil
	}
	if dailyTraffic == nil {
		dailyTraffic = nil
	}
	if onlineIPs == nil {
		onlineIPs = nil
	}

	Success(c, gin.H{
		"user":          user,
		"node_traffic":  nodeTraffic,
		"daily_traffic": dailyTraffic,
		"online_ips":    onlineIPs,
	})
}

// parseInt is a helper to parse a query parameter as int.
func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}