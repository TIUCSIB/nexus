package service

import (
	"errors"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"
)

type SubscriptionInfo struct {
	User     model.User  `json:"user"`
	Plan     *model.Plan `json:"plan,omitempty"`
	IsActive bool        `json:"is_active"`
	Expired  bool        `json:"expired"`
	DaysLeft int         `json:"days_left"`
}

var (
	ErrUserNotFound         = errors.New("用户不存在")
	ErrUserDisabled         = errors.New("用户已被禁用")
	ErrUserExpired          = errors.New("用户订阅已过期")
	ErrUserTrafficExhausted = errors.New("流量已用尽")
)

// CheckUserSubscriptionAvailable verifies whether a user is allowed to fetch
// subscription content.
func CheckUserSubscriptionAvailable(user *model.User) error {
	if user == nil {
		return ErrUserNotFound
	}
	if user.Status != 1 {
		return ErrUserDisabled
	}
	if user.ExpiredAt != nil && !user.ExpiredAt.After(time.Now()) {
		return ErrUserExpired
	}
	if user.TrafficLimit > 0 && user.TrafficUsed >= user.TrafficLimit {
		return ErrUserTrafficExhausted
	}
	return nil
}

// SubscriptionUnavailableReason returns a user-facing reason for a failed
// subscription availability check.
func SubscriptionUnavailableReason(err error) string {
	switch {
	case errors.Is(err, ErrUserDisabled):
		return "账号已被禁用"
	case errors.Is(err, ErrUserExpired):
		return "订阅已过期"
	case errors.Is(err, ErrUserTrafficExhausted):
		return "流量已用尽"
	default:
		return "订阅不可用"
	}
}

// GetUserSubscriptionInfo 根据 Token 查询用户的订阅信息。
// 返回用户详情、关联套餐，以及是否有效的状态判断。
func GetUserSubscriptionInfo(token string) (*SubscriptionInfo, error) {
	if token == "" {
		return nil, ErrUserNotFound
	}

	var user model.User
	if err := database.DB.Where("token = ?", token).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	info := &SubscriptionInfo{
		User: user,
	}

	// 查询关联套餐
	if user.PlanID != nil && *user.PlanID > 0 {
		var plan model.Plan
		if err := database.DB.First(&plan, *user.PlanID).Error; err == nil {
			info.Plan = &plan
		}
	}

	if err := CheckUserSubscriptionAvailable(&user); err != nil {
		info.IsActive = false
		if errors.Is(err, ErrUserExpired) {
			info.Expired = true
			info.DaysLeft = 0
		}
		return info, err
	}

	info.IsActive = true
	if user.ExpiredAt != nil {
		info.DaysLeft = int(time.Until(*user.ExpiredAt).Hours() / 24)
		if info.DaysLeft < 0 {
			info.DaysLeft = 0
		}
	}

	return info, nil
}
