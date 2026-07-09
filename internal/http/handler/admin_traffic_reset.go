package handler

import (
	"fmt"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminTrafficResetUsers 返回所有用户的流量重置信息
// GET /api/admin/traffic-reset/users
func AdminTrafficResetUsers(c *gin.Context) {
	page, pageSize := parsePagination(c)
	q := c.Query("q")

	query := database.DB.Model(&model.User{}).
		Select("users.*, COALESCE(plans.name, '') as plan_name, COALESCE(plans.traffic_reset, 0) as plan_traffic_reset").
		Joins("LEFT JOIN plans ON plans.id = users.plan_id")

	if q != "" {
		query = query.Where("users.email LIKE ?", "%"+q+"%")
	}

	var total int64
	query.Count(&total)

	type trafficResetUser struct {
		ID                uint       `json:"id"`
		Email             string     `json:"email"`
		PlanID            *uint      `json:"plan_id"`
		PlanName          string     `json:"plan_name"`
		PlanTrafficReset  int        `json:"plan_traffic_reset"`
		TrafficUsed       int64      `json:"traffic_used"`
		TrafficLimit      int64      `json:"traffic_limit"`
		TrafficResetAt    *time.Time `json:"traffic_reset_at"`
		ExpiredAt         *time.Time `json:"expired_at"`
		Status            int        `json:"status"`
	}

	var results []trafficResetUser
	database.DB.Raw(`
		SELECT u.id, u.email, u.plan_id, COALESCE(p.name, '') as plan_name,
			COALESCE(p.traffic_reset, 0) as plan_traffic_reset,
			u.traffic_used, u.traffic_limit, u.traffic_reset_at, u.expired_at, u.status
		FROM users u
		LEFT JOIN plans p ON p.id = u.plan_id
		WHERE (? = '' OR u.email LIKE ?)
		ORDER BY u.id DESC
		LIMIT ? OFFSET ?
	`, q, "%"+q+"%", pageSize, (page-1)*pageSize).Scan(&results)

	if results == nil {
		results = []trafficResetUser{}
	}

	SuccessPage(c, results, total, page, pageSize)
}

// AdminManualTrafficReset 手动触发全部流量重置
// POST /api/admin/traffic-reset/manual
func AdminManualTrafficReset(c *gin.Context) {
	adminID := c.GetUint("user_id")
	// 查找管理员邮箱
	var adminUser model.User
	adminEmail := "admin"
	if err := database.DB.First(&adminUser, adminID).Error; err == nil {
		adminEmail = adminUser.Email
	}
	now := time.Now()

	// 重置所有有效用户的流量
	var users []model.User
	database.DB.Where("status = 1").Find(&users)

	resetCount := 0
	for _, user := range users {
		// 记录重置日志
		planName := ""
		var plan model.Plan
		if user.PlanID != nil && *user.PlanID > 0 {
			if err := database.DB.First(&plan, *user.PlanID).Error; err == nil {
				planName = plan.Name
			}
		}

		database.DB.Create(&model.TrafficResetLog{
			UserID:    user.ID,
			UserEmail: user.Email,
			PlanID:    user.PlanID,
			PlanName:  planName,
			Method:    "manual",
			Operator:  adminEmail,
			CreatedAt: now,
		})

		resetCount++
	}

	// 批量更新流量
	result := database.DB.Model(&model.User{}).Where("status = 1").Updates(map[string]interface{}{
		"traffic_used":     0,
		"upload_used":      0,
		"download_used":    0,
		"traffic_reset_at": &now,
	})
	if result.Error != nil {
		InternalError(c, "批量重置流量失败")
		return
	}

	notifyAllAgentsReload()

	detail := fmt.Sprintf("手动重置 %d 个用户的流量", resetCount)
	recordAudit(c, "manual_traffic_reset", detail, "")

	Success(c, gin.H{
		"message":     fmt.Sprintf("已重置 %d 个用户的流量", resetCount),
		"reset_count": resetCount,
	})
}

// AdminTrafficResetStats 返回流量重置统计
// GET /api/admin/traffic-reset/stats
func AdminTrafficResetStats(c *gin.Context) {
	todayStart := time.Now().Truncate(24 * time.Hour)
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())

	var todayResetCount int64
	database.DB.Model(&model.TrafficResetLog{}).Where("created_at >= ?", todayStart).Count(&todayResetCount)

	var monthResetCount int64
	database.DB.Model(&model.TrafficResetLog{}).Where("created_at >= ?", monthStart).Count(&monthResetCount)

	var totalResetCount int64
	database.DB.Model(&model.TrafficResetLog{}).Count(&totalResetCount)

	// 各操作者统计
	var operatorStats []struct {
		Operator string `json:"operator"`
		Count    int64  `json:"count"`
	}
	database.DB.Model(&model.TrafficResetLog{}).
		Select("operator, COUNT(*) as count").
		Group("operator").
		Scan(&operatorStats)
	if operatorStats == nil {
		operatorStats = []struct {
			Operator string `json:"operator"`
			Count    int64  `json:"count"`
		}{}
	}

	Success(c, gin.H{
		"today_reset":  todayResetCount,
		"month_reset":  monthResetCount,
		"total_reset":  totalResetCount,
		"by_operator":  operatorStats,
	})
}