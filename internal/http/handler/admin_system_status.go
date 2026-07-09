package handler

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// AdminSystemStatus 返回面板系统状态
// GET /api/admin/stats/system
func AdminSystemStatus(c *gin.Context) {
	// Go 版本和运行时间
	goVersion := runtime.Version()
	uptime := time.Since(startTime).String()

	// 数据库文件大小
	dbSize := int64(0)
	if info, err := os.Stat("data/nexus.db"); err == nil {
		dbSize = info.Size()
	}

	// 统计信息
	var totalUsers, activeUsers int64
	database.DB.Model(&model.User{}).Count(&totalUsers)
	database.DB.Model(&model.User{}).Where("status = 1").Count(&activeUsers)

	var totalNodes, onlineNodes int64
	database.DB.Model(&model.Node{}).Count(&totalNodes)
	database.DB.Model(&model.Node{}).Where("online = ?", true).Count(&onlineNodes)

	var onlineDevices int64
	database.DB.Model(&model.AliveIP{}).Where("updated_at >= ?", time.Now().Add(-3*time.Minute)).Count(&onlineDevices)

	var onlineUsers int64
	database.DB.Model(&model.AliveIP{}).
		Select("COUNT(DISTINCT user_id)").
		Where("updated_at >= ?", time.Now().Add(-3*time.Minute)).
		Scan(&onlineUsers)

	// 今日流量
	todayStart := time.Now().Truncate(24 * time.Hour)
	var todayTraffic int64
	database.DB.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(upload + download), 0)").
		Where("recorded_at >= ?", todayStart).
		Scan(&todayTraffic)

	// 系统版本
	version := "1.0.0"
	if v := os.Getenv("NEXUS_VERSION"); v != "" {
		version = v
	}

	Success(c, gin.H{
		"version":        version,
		"go_version":     goVersion,
		"uptime":         uptime,
		"db_size":        dbSize,
		"db_size_human":  formatBytes(dbSize),
		"total_users":    totalUsers,
		"active_users":   activeUsers,
		"total_nodes":    totalNodes,
		"online_nodes":   onlineNodes,
		"online_devices": onlineDevices,
		"online_users":   onlineUsers,
		"today_traffic":  todayTraffic,
		"start_time":     startTime.Format("2006-01-02 15:04:05"),
	})
}

func formatBytes(b int64) string {
	if b == 0 {
		return "0 B"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}