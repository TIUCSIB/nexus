package handler

import (
	"fmt"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

type groupWithStats struct {
	model.ServerGroup
	UserCount  int64 `json:"user_count"`
	NodeCount  int64 `json:"node_count"`
}

func AdminListGroups(c *gin.Context) {
	var groups []model.ServerGroup
	database.DB.Order("id ASC").Find(&groups)

	var result []groupWithStats
	for _, g := range groups {
var userCount, nodeCount int64
			database.DB.Model(&model.User{}).Where("group_id = ?", g.ID).Count(&userCount)
			database.DB.Model(&model.Node{}).Where("group_id = ? OR group_ids LIKE ?", g.ID, fmt.Sprintf("%%%d%%", g.ID)).Count(&nodeCount)
		result = append(result, groupWithStats{ServerGroup: g, UserCount: userCount, NodeCount: nodeCount})
	}

	Success(c, result)
}

func AdminCreateGroup(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请输入权限组名称")
		return
	}

	group := model.ServerGroup{Name: req.Name}
	if err := database.DB.Create(&group).Error; err != nil {
		InternalError(c, "创建失败，名称可能重复")
		return
	}
	Success(c, group)
}

func AdminUpdateGroup(c *gin.Context) {
	id := c.Param("id")
	var group model.ServerGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		NotFound(c, "权限组不存在")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请求参数格式错误")
		return
	}

	if req.Name != "" {
		database.DB.Model(&group).Update("name", req.Name)
	}

	database.DB.First(&group, id)
	Success(c, group)
}

func AdminDeleteGroup(c *gin.Context) {
	id := c.Param("id")
	var group model.ServerGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		NotFound(c, "权限组不存在")
		return
	}

	var nodeCount, planCount int64
	database.DB.Model(&model.Node{}).Where("group_id = ? OR group_ids LIKE ?", group.ID, fmt.Sprintf("%%%d%%", group.ID)).Count(&nodeCount)
	database.DB.Model(&model.Plan{}).Where("group_id = ?", group.ID).Count(&planCount)

	if nodeCount > 0 || planCount > 0 {
		BadRequest(c, "该权限组下还有节点或套餐，无法删除")
		return
	}

	database.DB.Delete(&group)
	Success(c, gin.H{"message": "权限组已删除"})
}
