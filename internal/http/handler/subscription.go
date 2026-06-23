package handler

import (
	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

func lookupUserByToken(c *gin.Context) (*model.User, bool) {
	token := c.Query("token")
	if token == "" {
		BadRequest(c, "缺少订阅令牌参数")
		return nil, false
	}

	var user model.User
	if err := database.DB.Where("token = ?", token).First(&user).Error; err != nil {
		Unauthorized(c, "订阅令牌无效")
		return nil, false
	}

	if user.Status != 1 {
		Forbidden(c, "账号已被禁用")
		return nil, false
	}

	return &user, true
}

func SubSingbox(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}

	Success(c, gin.H{
		"format": "singbox",
		"token":  user.Token,
		"status": "待实现 - sing-box 订阅生成",
	})
}

func SubClash(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}

	Success(c, gin.H{
		"format": "clash",
		"token":  user.Token,
		"status": "待实现 - Clash 订阅生成",
	})
}

func SubSurge(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}

	Success(c, gin.H{
		"format": "surge",
		"token":  user.Token,
		"status": "待实现 - Surge 订阅生成",
	})
}

func SubSurfboard(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}

	Success(c, gin.H{
		"format": "surfboard",
		"token":  user.Token,
		"status": "待实现 - Surfboard 订阅生成",
	})
}

func SubShadowrocket(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}

	Success(c, gin.H{
		"format": "shadowrocket",
		"token":  user.Token,
		"status": "待实现 - Shadowrocket 订阅生成",
	})
}

func SubV2RayN(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}

	Success(c, gin.H{
		"format": "v2rayn",
		"token":  user.Token,
		"status": "待实现 - V2RayN 订阅生成",
	})
}
