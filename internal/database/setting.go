package database

import (
	"strconv"
	"strings"

	"nexus/internal/model"
)

// GetSetting 读取 system_configs 表中某个 key 的原始字符串值。
func GetSetting(key string) string {
	if DB == nil {
		return ""
	}
	var cfg model.SystemConfig
	if err := DB.Where("`key` = ?", key).First(&cfg).Error; err != nil {
		return ""
	}
	return cfg.Value
}

// GetSettingDefault 读取某个 key，不存在或为空时返回默认值。
func GetSettingDefault(key, def string) string {
	if v := GetSetting(key); v != "" {
		return v
	}
	return def
}

// GetSettingBool 读取某个布尔型设置，支持 "true"/"1"/"yes" 为真。
func GetSettingBool(key string, def bool) bool {
	v := GetSetting(key)
	if v == "" {
		return def
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return def
	}
}

// GetSettingInt 读取某个整型设置，解析失败或不存在时返回默认值。
func GetSettingInt(key string, def int) int {
	v := GetSetting(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return n
}