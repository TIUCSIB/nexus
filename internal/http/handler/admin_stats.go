package handler

import (
	"strconv"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

func AdminStatsOverview(c *gin.Context) {
	var totalUsers int64
	database.DB.Model(&model.User{}).Count(&totalUsers)

	var totalNodes int64
	database.DB.Model(&model.Node{}).Count(&totalNodes)

	var onlineNodes int64
	database.DB.Model(&model.Node{}).Where("online = ?", true).Count(&onlineNodes)

	var totalUpload int64
	var totalDownload int64
	database.DB.Model(&model.TrafficLog{}).Select("COALESCE(SUM(upload), 0)").Scan(&totalUpload)
	database.DB.Model(&model.TrafficLog{}).Select("COALESCE(SUM(download), 0)").Scan(&totalDownload)

	// Today's traffic
	todayStart := time.Now().Truncate(24 * time.Hour)
	var todayUpload int64
	var todayDownload int64
	database.DB.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(upload), 0)").
		Where("recorded_at >= ?", todayStart).
		Scan(&todayUpload)
	database.DB.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(download), 0)").
		Where("recorded_at >= ?", todayStart).
		Scan(&todayDownload)

	// 在线设备数
	var onlineDevices int64
	database.DB.Model(&model.AliveIP{}).Where("updated_at >= ?", time.Now().Add(-3*time.Minute)).Count(&onlineDevices)

	// 在线用户数
	var onlineUsers int64
	database.DB.Model(&model.AliveIP{}).
		Select("COUNT(DISTINCT user_id)").
		Where("updated_at >= ?", time.Now().Add(-3*time.Minute)).
		Scan(&onlineUsers)

	// 月度流量
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	var monthlyTraffic int64
	database.DB.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(upload + download), 0)").
		Where("recorded_at >= ?", monthStart).
		Scan(&monthlyTraffic)

	// 昨日节点排行
	yesterdayStart := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -1)
	yesterdayEnd := time.Now().Truncate(24 * time.Hour)
	type yesterdayNodeRanking struct {
		NodeID   uint   `json:"node_id"`
		Name     string `json:"name"`
		Total    int64  `json:"total"`
		Upload   int64  `json:"upload"`
		Download int64  `json:"download"`
	}
	var yesterdayRanking []yesterdayNodeRanking
	database.DB.Table("traffic_logs tl").
		Select("tl.node_id, COALESCE(n.name, '') as name, SUM(tl.upload + tl.download) as total, SUM(tl.upload) as upload, SUM(tl.download) as download").
		Joins("LEFT JOIN nodes n ON n.id = tl.node_id").
		Where("tl.recorded_at >= ? AND tl.recorded_at < ?", yesterdayStart, yesterdayEnd).
		Group("tl.node_id").
		Order("total DESC").
		Limit(10).
		Scan(&yesterdayRanking)
	if yesterdayRanking == nil {
		yesterdayRanking = []yesterdayNodeRanking{}
	}

	Success(c, gin.H{
		"total_users":       totalUsers,
		"total_nodes":       totalNodes,
		"online_nodes":      onlineNodes,
		"total_traffic":     totalUpload + totalDownload,
		"total_upload":      totalUpload,
		"total_download":    totalDownload,
		"today_upload":      todayUpload,
		"today_download":    todayDownload,
		"today_traffic":     todayUpload + todayDownload,
		"online_devices":    onlineDevices,
		"online_users":      onlineUsers,
		"monthly_traffic":   monthlyTraffic,
		"yesterday_ranking": yesterdayRanking,
	})
}

type trafficByDay struct {
	Date     string `json:"date"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Total    int64  `json:"total"`
}

func AdminStatsTraffic(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	startDate := time.Now().AddDate(0, 0, -days)

	var results []trafficByDay
	database.DB.Model(&model.TrafficLog{}).
		Select("DATE(recorded_at) as date, SUM(upload) as upload, SUM(download) as download, SUM(upload + download) as total").
		Where("recorded_at >= ?", startDate).
		Group("DATE(recorded_at)").
		Order("date ASC").
		Scan(&results)

	if results == nil {
		results = []trafficByDay{}
	}

	Success(c, gin.H{
		"days":    days,
		"records": results,
	})
}
