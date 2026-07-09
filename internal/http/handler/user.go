package handler

import (
	"fmt"
	"strings"
	"time"

	"nexus/internal/config"
	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/pkg/crypto"
	"nexus/internal/service"

	"github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		NotFound(c, "用户不存在")
		return
	}

	Success(c, user)
}

// GetUserStats 返回当前用户的流量统计
// GET /api/user/stats
func GetUserStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		NotFound(c, "用户不存在")
		return
	}

	// 今日流量
	todayStart := time.Now().Truncate(24 * time.Hour)
	var todayUpload, todayDownload int64
	database.DB.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(upload), 0)").
		Where("user_id = ? AND recorded_at >= ?", userID, todayStart).
		Scan(&todayUpload)
	database.DB.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(download), 0)").
		Where("user_id = ? AND recorded_at >= ?", userID, todayStart).
		Scan(&todayDownload)

	// 本月流量
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	var monthlyUpload, monthlyDownload int64
	database.DB.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(upload), 0)").
		Where("user_id = ? AND recorded_at >= ?", userID, monthStart).
		Scan(&monthlyUpload)
	database.DB.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(download), 0)").
		Where("user_id = ? AND recorded_at >= ?", userID, monthStart).
		Scan(&monthlyDownload)

	// 各节点流量分布
	var nodeTraffic []struct {
		NodeID   uint   `json:"node_id"`
		NodeName string `json:"node_name"`
		Upload   int64  `json:"upload"`
		Download int64  `json:"download"`
	}
	database.DB.Table("traffic_logs t").
		Select("t.node_id, n.name as node_name, SUM(t.upload) as upload, SUM(t.download) as download").
		Joins("LEFT JOIN nodes n ON n.id = t.node_id").
		Where("t.user_id = ?", userID).
		Group("t.node_id").
		Scan(&nodeTraffic)
	if nodeTraffic == nil {
		nodeTraffic = []struct {
			NodeID   uint   `json:"node_id"`
			NodeName string `json:"node_name"`
			Upload   int64  `json:"upload"`
			Download int64  `json:"download"`
		}{}
	}

	// 最近30日每日流量
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var dailyTraffic []struct {
		Date     string `json:"date"`
		Upload   int64  `json:"upload"`
		Download int64  `json:"download"`
	}
	database.DB.Table("traffic_logs").
		Select("DATE(recorded_at) as date, SUM(upload) as upload, SUM(download) as download").
		Where("user_id = ? AND recorded_at >= ?", userID, thirtyDaysAgo).
		Group("DATE(recorded_at)").
		Order("date ASC").
		Scan(&dailyTraffic)
	if dailyTraffic == nil {
		dailyTraffic = []struct {
			Date     string `json:"date"`
			Upload   int64  `json:"upload"`
			Download int64  `json:"download"`
		}{}
	}

	Success(c, gin.H{
		"total_traffic":    user.TrafficUsed,
		"total_upload":     user.UploadUsed,
		"total_download":   user.DownloadUsed,
		"today_upload":     todayUpload,
		"today_download":   todayDownload,
		"monthly_upload":   monthlyUpload,
		"monthly_download": monthlyDownload,
		"node_traffic":     nodeTraffic,
		"daily_traffic":    dailyTraffic,
	})
}

type updateProfileRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请求参数格式错误")
		return
	}

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		NotFound(c, "用户不存在")
		return
	}

	updates := map[string]interface{}{}

	if req.Email != "" && req.Email != user.Email {
		var count int64
		database.DB.Model(&model.User{}).Where("email = ? AND id != ?", req.Email, userID).Count(&count)
		if count > 0 {
			BadRequest(c, "该邮箱已被其他账号使用")
			return
		}
		updates["email"] = req.Email
	}

	if req.Password != "" {
		if len(req.Password) < 8 {
			BadRequest(c, "密码长度不能少于8位")
			return
		}
		hash, err := crypto.HashPassword(req.Password)
		if err != nil {
			InternalError(c, "密码加密失败")
			return
		}
		updates["password_hash"] = hash
	}

	if len(updates) == 0 {
		BadRequest(c, "没有需要更新的字段")
		return
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		InternalError(c, "更新个人资料失败")
		return
	}

	database.DB.First(&user, userID)
	Success(c, user)
}

func getSubBaseURL(c *gin.Context) string {
	if subURL := database.GetSetting("sub_url"); subURL != "" {
		return subURL
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	if host != "" {
		return scheme + "://" + host
	}

	return fmt.Sprintf("http://%s:%d", config.Global.Server.Host, config.Global.Server.Port)
}

func getSubPath() string {
	p := strings.Trim(database.GetSettingDefault("sub_path", "s"), "/")
	if p == "" {
		p = "s"
	}
	return p
}

func GetSubscription(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		NotFound(c, "用户不存在")
		return
	}

	baseURL := getSubBaseURL(c)
	token := user.Token
	available := true
	unavailableReason := ""
	if err := service.CheckUserSubscriptionAvailable(&user); err != nil {
		available = false
		unavailableReason = service.SubscriptionUnavailableReason(err)
	}

	subURLs := strings.Split(baseURL, ",")
	subSeg := getSubPath()
	links := []string{}
	if available {
		for _, url := range subURLs {
			url = strings.TrimSpace(url)
			if url == "" {
				continue
			}
			links = append(links,
				url+"/api/"+subSeg+"/singbox?token="+token,
				url+"/api/"+subSeg+"/clash?token="+token,
				url+"/api/"+subSeg+"/surge?token="+token,
				url+"/api/"+subSeg+"/surfboard?token="+token,
				url+"/api/"+subSeg+"/shadowrocket?token="+token,
				url+"/api/"+subSeg+"/v2rayn?token="+token,
			)
		}
	}

	// 干净格式链接：/{sub_path}/{token}（根据客户端自动识别格式）
	cleanLinks := []string{}
	if available {
		for _, url := range subURLs {
			url = strings.TrimSpace(url)
			if url == "" {
				continue
			}
			cleanLinks = append(cleanLinks, url+"/"+subSeg+"/"+token)
		}
	}

	var planName string
	if user.PlanID != nil && *user.PlanID > 0 {
		var plan model.Plan
		if err := database.DB.First(&plan, *user.PlanID).Error; err == nil {
			planName = plan.Name
		}
	}

	Success(c, gin.H{
		"plan_id":            user.PlanID,
		"plan_name":          planName,
		"traffic_used":       user.TrafficUsed,
		"traffic_limit":      user.TrafficLimit,
		"expired_at":         user.ExpiredAt,
		"token":              token,
		"links":              links,
		"clean_links":        cleanLinks,
		"sub_path":           subSeg,
		"available":          available,
		"unavailable_reason": unavailableReason,
	})
}
