package handler

import (
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminListAuditLogs 返回操作审计日志
// GET /api/admin/audit-logs
func AdminListAuditLogs(c *gin.Context) {
	page, pageSize := parsePagination(c)
	action := c.Query("action")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := database.DB.Model(&model.AuditLog{})

	if action != "" {
		query = query.Where("action LIKE ?", "%"+action+"%")
	}

	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}

	var total int64
	query.Count(&total)

	var logs []model.AuditLog
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)

	if logs == nil {
		logs = []model.AuditLog{}
	}

	SuccessPage(c, logs, total, page, pageSize)
}