package handler

import (
	"fmt"
	"strings"

	"nexus/internal/config"
	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/pkg/crypto"

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

	subURLs := strings.Split(baseURL, ",")
	subSeg := getSubPath()
	var links []string
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

	// 干净格式链接：/{sub_path}/{token}（默认 singbox 格式）
	var cleanLinks []string
	for _, url := range subURLs {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		cleanLinks = append(cleanLinks, url+"/"+subSeg+"/"+token)
	}

	var planName string
	if user.PlanID != nil && *user.PlanID > 0 {
		var plan model.Plan
		if err := database.DB.First(&plan, *user.PlanID).Error; err == nil {
			planName = plan.Name
		}
	}

	Success(c, gin.H{
		"plan_id":       user.PlanID,
		"plan_name":     planName,
		"traffic_used":  user.TrafficUsed,
		"traffic_limit": user.TrafficLimit,
		"expired_at":    user.ExpiredAt,
		"token":         token,
		"links":         links,
		"clean_links":   cleanLinks,
		"sub_path":      subSeg,
	})
}