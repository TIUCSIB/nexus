package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/service"
	"nexus/internal/subscription"

	"github.com/gin-gonic/gin"
)

// subFormatMap 订阅格式名称到处理函数的映射
var subFormatMap = map[string]func(*gin.Context){
	"singbox":      SubSingbox,
	"clash":        SubClash,
	"clashmeta":    SubClash,
	"surge":        SubSurge,
	"surfboard":    SubSurfboard,
	"shadowrocket": SubShadowrocket,
	"v2rayn":       SubV2RayN,
}

// setQueryToken writes token into the query string while preserving existing
// parameters such as format.
func setQueryToken(c *gin.Context, token string) {
	q := c.Request.URL.Query()
	q.Set("token", token)
	c.Request.URL.RawQuery = q.Encode()
}

// SubRouter 动态订阅路由处理器，根据当前 sub_path 设置实时分发请求
// 匹配模式：/api/{sub_path}/{format}、/api/{sub_path}/{token} 或 /api/{format}/{token}
func SubRouter(c *gin.Context) {
	subPath := strings.Trim(database.GetSettingDefault("sub_path", "s"), "/")
	if subPath == "" {
		subPath = "s"
	}

	segments := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
	if len(segments) < 3 || segments[0] != "api" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// /api/{sub_path}/{format} or /api/{sub_path}/{token}
	if segments[1] == subPath {
		segment := segments[2]
		if h, ok := subFormatMap[segment]; ok {
			h(c)
			return
		}

		setQueryToken(c, segment)
		SubAutoDetect(c)
		return
	}

	// /api/{format}/{token}
	if h, ok := subFormatMap[segments[1]]; ok {
		setQueryToken(c, segments[2])
		h(c)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

// SubAutoDetect 根据 format 参数或 User-Agent 自动选择订阅格式
func SubAutoDetect(c *gin.Context) {
	format := strings.ToLower(strings.TrimSpace(c.Query("format")))
	if h, ok := subFormatMap[format]; ok {
		h(c)
		return
	}

	ua := c.GetHeader("User-Agent")

	switch {
	case strings.Contains(ua, "ClashMeta") || strings.Contains(ua, "Mihomo") ||
		strings.Contains(ua, "Stash") || strings.Contains(ua, "verge") ||
		strings.Contains(ua, "flclash") || strings.Contains(ua, "nekobox") ||
		strings.Contains(ua, "clashmeta"):
		SubClash(c)
	case strings.Contains(ua, "Clash") || strings.Contains(ua, "clash"):
		SubClash(c)
	case strings.Contains(ua, "Surge"):
		SubSurge(c)
	case strings.Contains(ua, "Surfboard"):
		SubSurfboard(c)
	case strings.Contains(ua, "Shadowrocket"):
		SubShadowrocket(c)
	default:
		// V2RayN / 通用 Base64 格式
		SubV2RayN(c)
	}
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

	if err := service.CheckUserSubscriptionAvailable(&user); err != nil {
		Forbidden(c, service.SubscriptionUnavailableReason(err))
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
		// 支持多分组：匹配 group_id 或 group_ids JSON 中包含用户组
		groupIDStr := fmt.Sprintf("%d", *user.GroupID)
		q = q.Where("group_id = ? OR group_ids LIKE ?", *user.GroupID, fmt.Sprintf("%%%s%%", groupIDStr))
	}

	q.Order("sort ASC, id ASC").Find(&nodes)

	// 随机打乱节点顺序（同 Xboard 行为，防止客户端固定顺序）
	// Go 1.20+ 的 math/rand 全局源已自动随机初始化
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	return nodes
}

// setInfoNodes 向节点列表顶部插入信息节点（Xboard 风格）
// 复制第一个节点的所有配置参数，只修改节点名称为套餐到期和剩余流量信息
func setInfoNodes(nodes []model.Node, user *model.User) []model.Node {
	if !database.GetSettingBool("sub_show_info", true) {
		return nodes
	}
	if len(nodes) == 0 {
		return nodes
	}

	// 找第一个有效的节点作为信息节点模板
	var firstNode *model.Node
	for i := range nodes {
		switch strings.ToLower(nodes[i].Protocol) {
		case "vless", "hysteria2", "tuic":
			firstNode = &nodes[i]
			break
		}
	}
	if firstNode == nil {
		return nodes
	}

	expiryName, trafficName := subscription.GetInfoNodeNames(*user)

	// 构建信息节点列表（剩余流量在前，套餐到期在后，符合 Xboard 顺序）
	infoNodes := make([]model.Node, 0, len(nodes)+2)

	// 剩余流量节点
	trafficNode := *firstNode
	trafficNode.ID = 0
	trafficNode.Name = trafficName
	infoNodes = append(infoNodes, trafficNode)

	// 套餐到期节点
	expiryNode := *firstNode
	expiryNode.ID = 0
	expiryNode.Name = expiryName
	infoNodes = append(infoNodes, expiryNode)

	// 追加真实节点
	infoNodes = append(infoNodes, nodes...)

	return infoNodes
}

func SubSingbox(c *gin.Context) {
	user, ok := lookupUserByToken(c)
	if !ok {
		return
	}
	setUserinfoHeader(c, user)

	nodes := setInfoNodes(getUserNodes(user), user)
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

	nodes := setInfoNodes(getUserNodes(user), user)
	appName := database.GetSettingDefault("app_name", "Proxy")
	data, err := subscription.GenerateClash(nodes, *user, appName)
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

	nodes := setInfoNodes(getUserNodes(user), user)
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

	nodes := setInfoNodes(getUserNodes(user), user)
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

	nodes := setInfoNodes(getUserNodes(user), user)
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

	nodes := setInfoNodes(getUserNodes(user), user)
	data, err := subscription.GenerateV2RayN(nodes, *user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
}
