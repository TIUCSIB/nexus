package handler

import (
	"nexus/internal/database"

	"github.com/gin-gonic/gin"
)

func GetSiteInfo(c *gin.Context) {
	settings := make(map[string]string)
	settings["app_name"] = database.GetSettingDefault("app_name", "Nexus")
	settings["app_description"] = database.GetSettingDefault("app_description", "")
	settings["sub_url"] = database.GetSettingDefault("sub_url", "")
	settings["sub_path"] = database.GetSettingDefault("sub_path", "s")

	Success(c, settings)
}