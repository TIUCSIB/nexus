package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nexus/internal/config"
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

// AdminBackupDatabase exports the SQLite database as a downloadable file.
// POST /api/admin/settings/backup
func AdminBackupDatabase(c *gin.Context) {
	dsn := config.Global.Database.DSN
	if dsn == "" {
		InternalError(c, "数据库路径未配置")
		return
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(dsn)
	if err != nil {
		InternalError(c, "无法解析数据库路径")
		return
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		NotFound(c, "数据库文件不存在")
		return
	}

	// Create a backup using VACUUM INTO (produces a clean, compact copy)
	backupDir := filepath.Join(filepath.Dir(absPath), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		InternalError(c, "无法创建备份目录")
		return
	}

	backupPath := filepath.Join(backupDir, fmt.Sprintf("nexus_%s.db", time.Now().Format("20060102_150405")))
	result := database.DB.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath))
	if result.Error != nil {
		InternalError(c, "备份失败: "+result.Error.Error())
		return
	}

	// Send file as download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(backupPath)))
	c.Header("Content-Type", "application/octet-stream")
	c.File(backupPath)

	// Clean up old backups (keep last 10)
	entries, _ := os.ReadDir(backupDir)
	if len(entries) > 10 {
		for _, entry := range entries[:len(entries)-10] {
			os.Remove(filepath.Join(backupDir, entry.Name()))
		}
	}
}

// AdminBackupInfo returns backup directory status.
// GET /api/admin/settings/backup-info
func AdminBackupInfo(c *gin.Context) {
	dsn := config.Global.Database.DSN
	absPath, _ := filepath.Abs(dsn)
	backupDir := filepath.Join(filepath.Dir(absPath), "backups")

	var backups []gin.H
	entries, err := os.ReadDir(backupDir)
	if err == nil {
		for _, entry := range entries {
			info, err := entry.Info()
			if err == nil {
				backups = append(backups, gin.H{
					"name": entry.Name(),
					"size": info.Size(),
					"time": info.ModTime(),
				})
			}
		}
	}

	if backups == nil {
		backups = []gin.H{}
	}

	Success(c, gin.H{
		"backup_dir": backupDir,
		"backups":    backups,
	})
}
