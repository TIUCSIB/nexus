package handler

import (
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AdminListNodes(c *gin.Context) {
	page, pageSize := parsePagination(c)

	var total int64
	database.DB.Model(&model.Node{}).Count(&total)

	var nodes []model.Node
	offset := (page - 1) * pageSize
	database.DB.Order("sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&nodes)

	SuccessPage(c, nodes, total, page, pageSize)
}

type createNodeRequest struct {
	Name       string `json:"name" binding:"required"`
	Address    string `json:"address" binding:"required"`
	Protocol   string `json:"protocol" binding:"required"`
	Port       int    `json:"port" binding:"required"`
	ConfigMode string `json:"config_mode"`
	ConfigJSON string `json:"config_json"`
	Sort       int    `json:"sort"`
	Status     *int   `json:"status"`
}

func AdminCreateNode(c *gin.Context) {
	var req createNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请输入节点名称、地址、协议和端口")
		return
	}

	configMode := req.ConfigMode
	if configMode == "" {
		configMode = "auto"
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	node := model.Node{
		Name:          req.Name,
		Address:       req.Address,
		Protocol:      req.Protocol,
		Port:          req.Port,
		ConfigMode:    configMode,
		ConfigJSON:    req.ConfigJSON,
		RegisterToken: uuid.New().String(),
		Sort:          req.Sort,
		Status:        status,
	}

	if err := database.DB.Create(&node).Error; err != nil {
		InternalError(c, "创建节点失败")
		return
	}

	Success(c, gin.H{
		"node":            node,
		"register_token":  node.RegisterToken,
	})
}

type updateNodeRequest struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	Protocol   string `json:"protocol"`
	Port       *int   `json:"port"`
	ConfigMode string `json:"config_mode"`
	ConfigJSON string `json:"config_json"`
	Sort       *int   `json:"sort"`
	Status     *int   `json:"status"`
}

func AdminUpdateNode(c *gin.Context) {
	id := c.Param("id")

	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	var req updateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请求参数格式错误")
		return
	}

	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Protocol != "" {
		updates["protocol"] = req.Protocol
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.ConfigMode != "" {
		updates["config_mode"] = req.ConfigMode
	}
	if req.ConfigJSON != "" {
		updates["config_json"] = req.ConfigJSON
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		BadRequest(c, "没有需要更新的字段")
		return
	}

	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&node).Updates(updates).Error; err != nil {
		InternalError(c, "更新节点失败")
		return
	}

	database.DB.First(&node, id)
	Success(c, node)
}

func AdminDeleteNode(c *gin.Context) {
	id := c.Param("id")

	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	if err := database.DB.Delete(&node).Error; err != nil {
		InternalError(c, "删除节点失败")
		return
	}

	Success(c, gin.H{"message": "节点已删除"})
}

func AdminRestartNode(c *gin.Context) {
	id := c.Param("id")

	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	if node.Status != 1 {
		BadRequest(c, "节点未启用，无法重启")
		return
	}

	Success(c, gin.H{"message": "重启指令已发送", "node_id": node.ID})
}
