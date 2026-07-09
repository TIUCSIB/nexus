package service

import (
	"log"
	"strconv"
	"time"

	nexusconfig "nexus/internal/config"
	nexusdb "nexus/internal/database"
	nexusmodel "nexus/internal/model"
)

const (
	resetTrafficNone    = "0"
	resetTrafficMonthly = "1"
	resetTrafficCycle   = "2"
	resetTrafficYearly  = "3"
)

var schedulerStop = make(chan struct{})

func StartScheduler() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		runTrafficReset()
		checkNodeOffline()

		for {
			select {
			case <-ticker.C:
				runTrafficReset()
				checkNodeOffline()
			case <-schedulerStop:
				return
			}
		}
	}()
}

func StopScheduler() {
	select {
	case <-schedulerStop:
	default:
		close(schedulerStop)
	}
}

// runTrafficReset 遍历所有有效用户，根据其套餐的 traffic_reset 设置或全局设置执行流量重置。
// 优先级: 套餐 traffic_reset > 全局 reset_traffic_method
func runTrafficReset() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[scheduler] traffic reset panicked: %v", r)
		}
	}()

	globalMethod := nexusdb.GetSettingDefault("reset_traffic_method", resetTrafficNone)
	now := time.Now()

	// 查询所有有效用户
	var users []nexusmodel.User
	nexusdb.DB.Where("status = 1").Find(&users)

	var resetCount int
	for _, user := range users {
		method := globalMethod

		// 检查用户的套餐是否有独立的 traffic_reset 设置
		if user.PlanID != nil && *user.PlanID > 0 {
			var plan nexusmodel.Plan
			if err := nexusdb.DB.First(&plan, *user.PlanID).Error; err == nil {
				if plan.TrafficReset > 0 {
					// 套餐有独立设置，覆盖全局
					method = strconv.Itoa(plan.TrafficReset)
				} else {
					// 套餐明确设为 0（不重置），跳过此用户
					continue
				}
			}
		}

		if method == resetTrafficNone {
			continue
		}

		shouldReset := false

		switch method {
		case resetTrafficMonthly:
			// 每月1号重置：traffic_reset_at 早于当月1号则重置
			if now.Day() == 1 {
				boundary := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				if user.TrafficResetAt == nil || user.TrafficResetAt.Before(boundary) {
					shouldReset = true
				}
			}

		case resetTrafficYearly:
			// 每年1月1号重置
			if now.Month() == time.January && now.Day() == 1 {
				boundary := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
				if user.TrafficResetAt == nil || user.TrafficResetAt.Before(boundary) {
					shouldReset = true
				}
			}

		case resetTrafficCycle:
			// 按订阅周期重置：根据套餐 duration_days 计算
			if user.PlanID != nil && *user.PlanID > 0 {
				var plan nexusmodel.Plan
				if err := nexusdb.DB.First(&plan, *user.PlanID).Error; err == nil && plan.DurationDays > 0 {
					if user.ExpiredAt != nil {
						cycleStart := user.ExpiredAt.AddDate(0, 0, -plan.DurationDays)
						if user.TrafficResetAt == nil || user.TrafficResetAt.Before(cycleStart) {
							shouldReset = true
						}
					}
				}
			}
		}

		if shouldReset {
			resetUserTrafficWithLog(user, method, now)
			resetCount++
		}
	}

	if resetCount > 0 {
		log.Printf("[scheduler] reset traffic for %d users", resetCount)
	}
}

// resetUserTrafficWithLog 重置指定用户的流量并记录日志
func resetUserTrafficWithLog(user nexusmodel.User, method string, now time.Time) {
	planName := ""
	if user.PlanID != nil && *user.PlanID > 0 {
		var plan nexusmodel.Plan
		if err := nexusdb.DB.First(&plan, *user.PlanID).Error; err == nil {
			planName = plan.Name
		}
	}

	// 记录重置日志
	nexusdb.DB.Create(&nexusmodel.TrafficResetLog{
		UserID:    user.ID,
		UserEmail: user.Email,
		PlanID:    user.PlanID,
		PlanName:  planName,
		Method:    method,
		Operator:  "system",
		CreatedAt: now,
	})

	// 重置流量
	nexusdb.DB.Model(&nexusmodel.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]any{
			"traffic_used":     0,
			"upload_used":      0,
			"download_used":    0,
			"traffic_reset_at": now,
		})
}

// ResetUserTraffic 手动重置单个用户的流量
func ResetUserTraffic(userID uint) error {
	now := time.Now()

	var user nexusmodel.User
	if err := nexusdb.DB.First(&user, userID).Error; err != nil {
		return err
	}

	planName := ""
	var plan nexusmodel.Plan
	if user.PlanID != nil && *user.PlanID > 0 {
		if err := nexusdb.DB.First(&plan, *user.PlanID).Error; err == nil {
			planName = plan.Name
		}
	}

	// 记录重置日志
	nexusdb.DB.Create(&nexusmodel.TrafficResetLog{
		UserID:    user.ID,
		UserEmail: user.Email,
		PlanID:    user.PlanID,
		PlanName:  planName,
		Method:    "manual",
		Operator:  "admin",
		CreatedAt: now,
	})

	return nexusdb.DB.Model(&nexusmodel.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"traffic_used":     0,
			"upload_used":      0,
			"download_used":    0,
			"traffic_reset_at": &now,
		}).Error
}

// checkNodeOffline marks nodes as offline if their last heartbeat exceeds the timeout.
func checkNodeOffline() {
	timeout := time.Duration(nexusconfig.Global.Node.OfflineTimeout) * time.Second
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	cutoff := time.Now().Add(-timeout)

	result := nexusdb.DB.Model(&nexusmodel.Node{}).
		Where("online = ? AND last_heartbeat IS NOT NULL AND last_heartbeat < ?", true, cutoff).
		Update("online", false)
	if result.Error != nil {
		log.Printf("[scheduler] failed to check node offline: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("[scheduler] marked %d nodes as offline", result.RowsAffected)
	}
}
