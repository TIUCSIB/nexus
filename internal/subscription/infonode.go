package subscription

import (
	"fmt"
	"math"
	"time"

	"nexus/internal/model"
)

// formatTraffic 将字节数格式化为可读的流量字符串
func formatTraffic(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	idx := 0
	size := float64(bytes)
	for size >= 1024 && idx < len(units)-1 {
		size /= 1024
		idx++
	}
	return fmt.Sprintf("%.2f %s", size, units[idx])
}

// formatExpiry 将到期时间格式化为可读字符串
func formatExpiry(expiredAt *time.Time) string {
	if expiredAt == nil {
		return "\u957f\u671f\u6709\u6548"
	}
	now := time.Now()
	if expiredAt.Before(now) {
		return "\u5df2\u8fc7\u671f"
	}
	days := int(math.Ceil(expiredAt.Sub(now).Hours() / 24))
	return fmt.Sprintf("%s (\u5269\u4f59 %d \u5929)", expiredAt.Format("2006-01-02"), days)
}

// GetInfoNodeNames 根据用户信息生成两个信息节点的名称
// 第一个节点显示到期时间，第二个节点显示剩余流量
func GetInfoNodeNames(user model.User) (string, string) {
	// 套餐到期
	expiryName := fmt.Sprintf("\u5957\u9910\u5230\u671f\uff1a%s", formatExpiry(user.ExpiredAt))

	// 剩余流量
	var remaining int64
	if user.TrafficLimit > 0 {
		remaining = user.TrafficLimit - user.TrafficUsed
		if remaining < 0 {
			remaining = 0
		}
	} else {
		// 无限制
		remaining = -1
	}

	var trafficName string
	if remaining < 0 {
		trafficName = "\u5269\u4f59\u6d41\u91cf\uff1a\u65e0\u9650"
	} else {
		trafficName = fmt.Sprintf("\u5269\u4f59\u6d41\u91cf\uff1a%s", formatTraffic(remaining))
	}

	return expiryName, trafficName
}
