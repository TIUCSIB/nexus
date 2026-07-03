package service

import (
	"log"
	"time"

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

		for {
			select {
			case <-ticker.C:
				runTrafficReset()
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

func runTrafficReset() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[scheduler] traffic reset panicked: %v", r)
		}
	}()

	method := nexusdb.GetSettingDefault("reset_traffic_method", resetTrafficNone)
	if method == resetTrafficNone {
		return
	}

	now := time.Now()

	switch method {
	case resetTrafficMonthly:
		// 每月1号重置：只在1号执行
		if now.Day() != 1 {
			return
		}
		boundary := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		resetUsersTraffic(boundary, method)

	case resetTrafficYearly:
		// 每年1月1号重置：只在1月1号执行
		if now.Month() != time.January || now.Day() != 1 {
			return
		}
		boundary := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
		resetUsersTraffic(boundary, method)

	case resetTrafficCycle:
		// 按订阅周期重置：根据每个用户的套餐时长判断
		resetUsersBySubscriptionCycle(now)
	}
}

// resetUsersTraffic 重置 traffic_reset_at 早于 boundary 的所有用户
func resetUsersTraffic(boundary time.Time, method string) {
	result := nexusdb.DB.Model(&nexusmodel.User{}).
		Where("traffic_reset_at IS NULL OR traffic_reset_at < ?", boundary).
		Updates(map[string]any{
			"traffic_used":     0,
			"traffic_reset_at": time.Now(),
		})
	if result.Error != nil {
		log.Printf("[scheduler] traffic reset failed (method=%s): %v", method, result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("[scheduler] reset traffic for %d users (method=%s)", result.RowsAffected, method)
	}
}

// resetUsersBySubscriptionCycle 按订阅周期重置：
// 对于每个用户，根据其套餐的 duration_days 计算上一个周期的起始时间，
// 如果 traffic_reset_at 早于该起始时间，则重置流量。
func resetUsersBySubscriptionCycle(now time.Time) {
	// 查询所有有效用户及其套餐
	var users []nexusmodel.User
	nexusdb.DB.Where("status = 1 AND plan_id IS NOT NULL").Find(&users)

	var resetCount int
	for _, user := range users {
		if user.PlanID == nil || *user.PlanID == 0 {
			continue
		}

		var plan nexusmodel.Plan
		if err := nexusdb.DB.First(&plan, *user.PlanID).Error; err != nil {
			continue
		}

		if plan.DurationDays <= 0 {
			// 套餐无期限限制，跳过
			continue
		}

		// 计算当前周期的起始时间
		// 基于 expired_at 往前推 duration_days 得到当前周期开始时间
		var cycleStart time.Time
		if user.ExpiredAt != nil {
			// 如果未过期，当前周期起始 = expired_at - duration_days
			cycleStart = user.ExpiredAt.AddDate(0, 0, -plan.DurationDays)
		} else {
			// 已过期，按过期时间计算
			// expired_at 可能为 nil，此时无法判断周期，跳过
			continue
		}

		// 如果用户上次重置时间早于当前周期起始时间，说明需要重置
		if user.TrafficResetAt == nil || user.TrafficResetAt.Before(cycleStart) {
			nexusdb.DB.Model(&nexusmodel.User{}).
				Where("id = ?", user.ID).
				Updates(map[string]any{
					"traffic_used":     0,
					"traffic_reset_at": now,
				})
			resetCount++
		}
	}

	if resetCount > 0 {
		log.Printf("[scheduler] reset traffic for %d users (method=cycle)", resetCount)
	}
}

// ResetUserTraffic 手动重置单个用户的流量
func ResetUserTraffic(userID uint) error {
	now := time.Now()
	return nexusdb.DB.Model(&nexusmodel.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"traffic_used":     0,
			"traffic_reset_at": &now,
		}).Error
}
