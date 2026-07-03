package handler

import (
	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

func AdminListOnlineIPs(c *gin.Context) {
	var results []struct {
		model.AliveIP
		UserEmail string `json:"user_email"`
		UserUUID  string `json:"user_uuid"`
		NodeName  string `json:"node_name"`
	}

	database.DB.Table("alive_ips a").
		Select("a.*, u.email as user_email, u.uuid as user_uuid, n.name as node_name").
		Joins("LEFT JOIN users u ON u.id = a.user_id").
		Joins("LEFT JOIN nodes n ON n.id = a.node_id").
		Order("a.updated_at DESC").
		Limit(200).
		Scan(&results)

	if results == nil {
		results = []struct {
			model.AliveIP
			UserEmail string `json:"user_email"`
			UserUUID  string `json:"user_uuid"`
			NodeName  string `json:"node_name"`
		}{}
	}

	Success(c, results)
}

func AdminListTrafficLogs(c *gin.Context) {
	page, pageSize := parsePagination(c)

	var total int64
	database.DB.Model(&model.TrafficLog{}).Count(&total)

	var logs []struct {
		model.TrafficLog
		UserEmail string `json:"user_email"`
		NodeName  string `json:"node_name"`
	}

	database.DB.Table("traffic_logs t").
		Select("t.*, u.email as user_email, n.name as node_name").
		Joins("LEFT JOIN users u ON u.id = t.user_id").
		Joins("LEFT JOIN nodes n ON n.id = t.node_id").
		Order("t.recorded_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&logs)

	if logs == nil {
		logs = []struct {
			model.TrafficLog
			UserEmail string `json:"user_email"`
			NodeName  string `json:"node_name"`
		}{}
	}

	SuccessPage(c, logs, total, page, pageSize)
}

func AdminGetUserOnlineIPs(c *gin.Context) {
	userID := c.Param("id")

	var results []struct {
		model.AliveIP
		NodeName string `json:"node_name"`
	}

	database.DB.Table("alive_ips a").
		Select("a.*, n.name as node_name").
		Joins("LEFT JOIN nodes n ON n.id = a.node_id").
		Where("a.user_id = ?", userID).
		Order("a.updated_at DESC").
		Limit(100).
		Scan(&results)

	if results == nil {
		results = []struct {
			model.AliveIP
			NodeName string `json:"node_name"`
		}{}
	}

	Success(c, results)
}

func AdminGetUserTrafficLogs(c *gin.Context) {
	userID := c.Param("id")

	var logs []struct {
		model.TrafficLog
		NodeName string `json:"node_name"`
	}

	database.DB.Table("traffic_logs t").
		Select("t.*, n.name as node_name").
		Joins("LEFT JOIN nodes n ON n.id = t.node_id").
		Where("t.user_id = ?", userID).
		Order("t.recorded_at DESC").
		Limit(100).
		Scan(&logs)

	if logs == nil {
		logs = []struct {
			model.TrafficLog
			NodeName string `json:"node_name"`
		}{}
	}

	Success(c, logs)
}
