package handler

import (
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

type createCustomOutboundRequest struct {
	Name         string `json:"name" binding:"required"`
	Tag          string `json:"tag" binding:"required"`
	Protocol     string `json:"protocol" binding:"required"`
	SettingsJSON string `json:"settings_json"`
	ProxyTag     string `json:"proxy_tag"`
	Sort         int    `json:"sort"`
	Status       *int   `json:"status"`
}

type updateCustomOutboundRequest struct {
	Name         string  `json:"name"`
	Tag          string  `json:"tag"`
	Protocol     string  `json:"protocol"`
	SettingsJSON *string `json:"settings_json"`
	ProxyTag     *string `json:"proxy_tag"`
	Sort         *int    `json:"sort"`
	Status       *int    `json:"status"`
}

type updateNodeOutboundsRequest struct {
	OutboundIDs []uint `json:"outbound_ids"`
}

func AdminListCustomOutbounds(c *gin.Context) {
	page, pageSize := parsePagination(c)
	var total int64
	query := database.DB.Model(&model.CustomOutbound{})
	query.Count(&total)

	var outbounds []model.CustomOutbound
	offset := (page - 1) * pageSize
	query.Order("sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&outbounds)
	SuccessPage(c, outbounds, total, page, pageSize)
}

func AdminCreateCustomOutbound(c *gin.Context) {
	var req createCustomOutboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请填写出站名称、标签和协议")
		return
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	outbound := model.CustomOutbound{
		Name:         req.Name,
		Tag:          req.Tag,
		Protocol:     req.Protocol,
		SettingsJSON: req.SettingsJSON,
		ProxyTag:     req.ProxyTag,
		Sort:         req.Sort,
		Status:       status,
	}
	if err := database.DB.Create(&outbound).Error; err != nil {
		InternalError(c, "创建自定义出站失败")
		return
	}
	Success(c, outbound)
}

func AdminUpdateCustomOutbound(c *gin.Context) {
	id := c.Param("id")
	var outbound model.CustomOutbound
	if err := database.DB.First(&outbound, id).Error; err != nil {
		NotFound(c, "自定义出站不存在")
		return
	}

	var req updateCustomOutboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请求参数格式错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Tag != "" {
		updates["tag"] = req.Tag
	}
	if req.Protocol != "" {
		updates["protocol"] = req.Protocol
	}
	if req.SettingsJSON != nil {
		updates["settings_json"] = *req.SettingsJSON
	}
	if req.ProxyTag != nil {
		updates["proxy_tag"] = *req.ProxyTag
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

	if err := database.DB.Model(&outbound).Updates(updates).Error; err != nil {
		InternalError(c, "更新自定义出站失败")
		return
	}
	database.DB.First(&outbound, id)
	Success(c, outbound)
}

func AdminDeleteCustomOutbound(c *gin.Context) {
	id := c.Param("id")
	var outbound model.CustomOutbound
	if err := database.DB.First(&outbound, id).Error; err != nil {
		NotFound(c, "自定义出站不存在")
		return
	}
	if err := database.DB.Delete(&outbound).Error; err != nil {
		InternalError(c, "删除自定义出站失败")
		return
	}
	database.DB.Where("custom_outbound_id = ?", outbound.ID).Delete(&model.NodeOutbound{})
	Success(c, gin.H{"message": "自定义出站已删除"})
}

func AdminListNodeOutbounds(c *gin.Context) {
	nodeID := c.Param("id")
	var node model.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	var bindings []model.NodeOutbound
	database.DB.Where("node_id = ?", node.ID).Find(&bindings)
	ids := make([]uint, 0, len(bindings))
	for _, b := range bindings {
		ids = append(ids, b.CustomOutboundID)
	}

	var outbounds []model.CustomOutbound
	if len(ids) > 0 {
		database.DB.Where("id IN ?", ids).Order("sort ASC, id ASC").Find(&outbounds)
	}
	Success(c, gin.H{"outbound_ids": ids, "outbounds": outbounds})
}

func AdminUpdateNodeOutbounds(c *gin.Context) {
	nodeID := c.Param("id")
	var node model.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		NotFound(c, "节点不存在")
		return
	}

	var req updateNodeOutboundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请求参数格式错误")
		return
	}

	tx := database.DB.Begin()
	if err := tx.Where("node_id = ?", node.ID).Delete(&model.NodeOutbound{}).Error; err != nil {
		tx.Rollback()
		InternalError(c, "清空节点出站绑定失败")
		return
	}
	for _, outboundID := range req.OutboundIDs {
		if outboundID == 0 {
			continue
		}
		binding := model.NodeOutbound{NodeID: node.ID, CustomOutboundID: outboundID}
		if err := tx.Create(&binding).Error; err != nil {
			tx.Rollback()
			InternalError(c, "保存节点出站绑定失败")
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		InternalError(c, "保存节点出站绑定失败")
		return
	}
	Success(c, gin.H{"message": "节点出站绑定已更新"})
}
