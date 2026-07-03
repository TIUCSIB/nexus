package handler

import (
	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

func AdminGetSettings(c *gin.Context) {
	var configs []model.SystemConfig
	database.DB.Find(&configs)

	settings := make(map[string]string, len(configs))
	for _, cfg := range configs {
		settings[cfg.Key] = cfg.Value
	}

	Success(c, settings)
}

type updateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

func AdminUpdateSettings(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请提供设置项")
		return
	}

	if len(req.Settings) == 0 {
		BadRequest(c, "设置项不能为空")
		return
	}

	for key, value := range req.Settings {
		config := model.SystemConfig{
			Key:   key,
			Value: value,
		}
		database.DB.Save(&config)
	}

	Success(c, gin.H{"message": "设置已保存"})
}
