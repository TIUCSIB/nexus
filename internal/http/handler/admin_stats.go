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

	Success(c, gin.H{
		"total_users":     totalUsers,
		"total_nodes":     totalNodes,
		"online_nodes":    onlineNodes,
		"total_traffic":   totalUpload + totalDownload,
		"total_upload":    totalUpload,
		"total_download":  totalDownload,
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
