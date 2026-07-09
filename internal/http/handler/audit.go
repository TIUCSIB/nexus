package handler

import (
	"encoding/json"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

// recordAudit 记录操作审计日志
func recordAudit(c *gin.Context, action, target, detail string) {
	userID := c.GetUint("user_id")
	userEmail, _ := c.Get("user_email")
	email, ok := userEmail.(string)
	if !ok || email == "" {
		// 查找用户邮箱
		var user model.User
		if err := database.DB.First(&user, userID).Error; err == nil {
			email = user.Email
		}
	}
	ip := c.ClientIP()

	database.DB.Create(&model.AuditLog{
		UserID:    userID,
		UserEmail: email,
		Action:    action,
		Target:    target,
		Detail:    detail,
		IP:        ip,
		CreatedAt: time.Now(),
	})
}

// recordAuditSimple 简化的审计记录函数，不依赖 gin.Context
func recordAuditSimple(userID uint, userEmail, action, target, detail, ip string) {
	database.DB.Create(&model.AuditLog{
		UserID:    userID,
		UserEmail: userEmail,
		Action:    action,
		Target:    target,
		Detail:    detail,
		IP:        ip,
		CreatedAt: time.Now(),
	})
}

// detailJSON 将任意数据转为 JSON 字符串，用于审计日志详情
func detailJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}