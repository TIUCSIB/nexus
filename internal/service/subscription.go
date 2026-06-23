package service

import (
	"errors"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"
)

type SubscriptionInfo struct {
	User       model.User `json:"user"`
	Plan       *model.Plan `json:"plan,omitempty"`
	IsActive   bool       `json:"is_active"`
	Expired    bool       `json:"expired"`
	DaysLeft   int        `json:"days_left"`
}

var (
	ErrUserNotFound  = errors.New("用户不存在")
	ErrUserDisabled  = errors.New("用户已被禁用")
	ErrUserExpired   = errors.New("用户订阅已过期")
)

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

	// 判断用户是否有效：状态为 1 且未过期
	if user.Status != 1 {
		info.IsActive = false
		return info, ErrUserDisabled
	}

	if user.ExpiredAt != nil && user.ExpiredAt.Before(time.Now()) {
		info.IsActive = false
		info.Expired = true
		info.DaysLeft = 0
		return info, ErrUserExpired
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
