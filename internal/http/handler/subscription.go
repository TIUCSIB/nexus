package handler

import (
	"fmt"
	"net/http"
	"strings"

	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/subscription"

	"github.com/gin-gonic/gin"
)

// subFormatMap 订阅格式名称到处理函数的映射
var subFormatMap = map[string]func(*gin.Context){
	"singbox":      SubSingbox,
	"clash":        SubClash,
	"surge":        SubSurge,
	"surfboard":    SubSurfboard,
	"shadowrocket": SubShadowrocket,
	"v2rayn":       SubV2RayN,
}

// SubRouter 动态订阅路由处理器，根据当前 sub_path 设置实时分发请求
// 匹配模式：/api/{sub_path}/{action} 或 /api/{sub_path}/{token}
func SubRouter(c *gin.Context) {
	subPath := strings.Trim(database.GetSettingDefault("sub_path", "s"), "/")
	if subPath == "" {
		subPath = "s"
	}

	// 从 URL 中提取第二个路径段（即 action 或 token）
	// 完整路径格式为 /api/{sub_path}/{segment}
	fullPath := c.Request.URL.Path
	segments := strings.Split(strings.Trim(fullPath, "/"), "/")

	// 期望格式：["api", "{sub_path}", "{segment}"]
	if len(segments) < 3 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// 检查第一个段是否为 api
	if segments[0] != "api" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// 检查第二个段是否匹配当前的 sub_path
	if segments[1] != subPath {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	segment := segments[2]

	// 先检查是否为已知的格式名称
	if handler, ok := subFormatMap[segment]; ok {
		handler(c)
		return
	}

	// 不是格式名称，则当作 token 处理（干净的订阅链接格式）
	// 将 token 放到 Query 参数中，复用 SubSingbox 处理器
	c.Request.URL.RawQuery = "token=" + segment
	SubSingbox(c)
}

func lookupUserByToken(c *gin.Context) (*model.User, bool) {
	token := c.Query("token")
	if token == "" {
		token = c.Param("token")
	}
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

func setUserinfoHeader(c *gin.Context, user *model.User) {
	if !database.GetSettingBool("sub_show_info", true) {
		return
	}

	used := user.TrafficUsed
	var total int64
	if user.TrafficLimit > 0 {
		total = user.TrafficLimit
	}

	header := fmt.Sprintf("upload=0; download=%d; total=%d", used, total)
	if user.ExpiredAt != nil {
		header += fmt.Sprintf("; expire=%d", user.ExpiredAt.Unix())
	}
	c.Header("Subscription-Userinfo", header)
}

func getUserNodes(user *model.User) []model.Node {
	var nodes []model.Node
	q := database.DB.Where("status = ?", 1)

	if user.GroupID != nil && *user.GroupID > 0 {
		q = q.Where("group_id = ?", *user.GroupID)
	}

	q.Order("sort ASC, id ASC").Find(&nodes)
	return nodes
}

func SubSingbox(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}
	setUserinfoHeader(c, user)

	nodes := getUserNodes(user)
	data, err := subscription.GenerateSingbox(nodes, *user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

func SubClash(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}
	setUserinfoHeader(c, user)

	nodes := getUserNodes(user)
	data, err := subscription.GenerateClash(nodes, *user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/yaml; charset=utf-8")
	c.Data(http.StatusOK, "text/yaml; charset=utf-8", data)
}

func SubSurge(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}
	setUserinfoHeader(c, user)

	nodes := getUserNodes(user)
	data, err := subscription.GenerateSurge(nodes, *user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
}

func SubSurfboard(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}
	setUserinfoHeader(c, user)

	nodes := getUserNodes(user)
	data, err := subscription.GenerateSurfboard(nodes, *user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
}

func SubShadowrocket(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}
	setUserinfoHeader(c, user)

	nodes := getUserNodes(user)
	data, err := subscription.GenerateShadowrocket(nodes, *user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
}

func SubV2RayN(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}
	setUserinfoHeader(c, user)

	nodes := getUserNodes(user)
	data, err := subscription.GenerateV2RayN(nodes, *user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
}