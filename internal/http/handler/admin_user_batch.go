package handler

import (
	"fmt"
	"strings"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminBatchUserOperation 批量用户操作
// POST /api/admin/users/batch-operation
func AdminBatchUserOperation(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Action string `json:"action" binding:"required"` // ban | unban | delete | reset_traffic | reset_uuid
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		BadRequest(c, "请提供用户ID列表和操作类型")
		return
	}

	// 验证操作类型
	validActions := map[string]bool{
		"ban":           true,
		"unban":         true,
		"delete":        true,
		"reset_traffic": true,
		"reset_uuid":    true,
	}
	if !validActions[req.Action] {
		BadRequest(c, "不支持的操作类型，支持: ban/unban/delete/reset_traffic/reset_uuid")
		return
	}

	// 禁止删除自己
	if req.Action == "delete" {
		adminID := c.GetUint("user_id")
		for _, id := range req.IDs {
			if id == adminID {
				BadRequest(c, "不能删除自己的账号")
				return
			}
		}
	}

	now := time.Now()

	switch req.Action {
	case "ban":
		database.DB.Model(&model.User{}).Where("id IN ?", req.IDs).Update("status", 0)
	case "unban":
		database.DB.Model(&model.User{}).Where("id IN ?", req.IDs).Update("status", 1)
	case "delete":
		// 删除用户及相关数据
		database.DB.Where("user_id IN ?", req.IDs).Delete(&model.AliveIP{})
		database.DB.Where("user_id IN ?", req.IDs).Delete(&model.TrafficLog{})
		database.DB.Where("id IN ?", req.IDs).Delete(&model.User{})
	case "reset_traffic":
		database.DB.Model(&model.User{}).Where("id IN ?", req.IDs).Updates(map[string]interface{}{
			"traffic_used":     0,
			"upload_used":      0,
			"download_used":    0,
			"traffic_reset_at": &now,
		})
	case "reset_uuid":
		// 为每个用户生成新的 UUID 和 token
		var users []model.User
		database.DB.Where("id IN ?", req.IDs).Find(&users)
		for _, u := range users {
			database.DB.Model(&model.User{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
				"uuid":  uuid.New().String(),
				"token": uuid.New().String(),
			})
		}
	}

	// 通知所有 Agent 重新加载
	notifyAllAgentsReload()

	// 记录审计日志
	actionLabels := map[string]string{
		"ban":           "批量封禁用户",
		"unban":         "批量解封用户",
		"delete":        "批量删除用户",
		"reset_traffic": "批量重置流量",
		"reset_uuid":    "批量重置UUID",
	}
	detail := fmt.Sprintf("操作: %s, 用户ID: %s", actionLabels[req.Action], joinUintIDs(req.IDs))
	recordAudit(c, "batch_"+req.Action, detail, "")

	Success(c, gin.H{"message": fmt.Sprintf("批量操作成功：%s", actionLabels[req.Action])})
}

func joinUintIDs(ids []uint) string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(s, ", ")
}