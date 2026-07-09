package handler

import (
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

// AdminListRoutes returns a paginated list of route rules.
// GET /api/v1/admin/routes
func AdminListRoutes(c *gin.Context) {
	page, pageSize := parsePagination(c)
	q := c.Query("q")

	query := database.DB.Model(&model.RouteRule{})
	if q != "" {
		query = query.Where("name LIKE ?", "%"+q+"%")
	}

	var total int64
	query.Count(&total)

	var rules []model.RouteRule
	offset := (page - 1) * pageSize
	query.Order("sort ASC, id DESC").Offset(offset).Limit(pageSize).Find(&rules)

	SuccessPage(c, rules, total, page, pageSize)
}

type createRouteRequest struct {
	Name        string `json:"name" binding:"required"`
	Match       string `json:"match"`
	Action      string `json:"action" binding:"required"`
	ActionValue string `json:"action_value"`
	MatchJSON   string `json:"match_json"`
	ActionJSON  string `json:"action_json"`
	Sort        int    `json:"sort"`
	Status      *int   `json:"status"`
}

// AdminCreateRoute creates a new route rule.
// POST /api/v1/admin/routes
func AdminCreateRoute(c *gin.Context) {
	var req createRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "\u8bf7\u8f93\u5165\u89c4\u5219\u540d\u79f0\u3001\u5339\u914d\u6761\u4ef6\u548c\u52a8\u4f5c")
		return
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	rule := model.RouteRule{
		Name:        req.Name,
		Match:       req.Match,
		Action:      req.Action,
		ActionValue: req.ActionValue,
		MatchJSON:   req.MatchJSON,
		ActionJSON:  req.ActionJSON,
		Sort:        req.Sort,
		Status:      status,
	}

	if err := database.DB.Create(&rule).Error; err != nil {
		InternalError(c, "\u521b\u5efa\u8def\u7531\u89c4\u5219\u5931\u8d25")
		return
	}

	Success(c, rule)
}

type updateRouteRequest struct {
	Name        *string `json:"name"`
	Match       *string `json:"match"`
	Action      *string `json:"action"`
	ActionValue *string `json:"action_value"`
	MatchJSON   *string `json:"match_json"`
	ActionJSON  *string `json:"action_json"`
	Sort        *int    `json:"sort"`
	Status      *int    `json:"status"`
}

// AdminUpdateRoute updates an existing route rule.
// PUT /api/v1/admin/routes/:id
func AdminUpdateRoute(c *gin.Context) {
	id := c.Param("id")

	var rule model.RouteRule
	if err := database.DB.First(&rule, id).Error; err != nil {
		NotFound(c, "\u8def\u7531\u89c4\u5219\u4e0d\u5b58\u5728")
		return
	}

	var req updateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "\u8bf7\u6c42\u53c2\u6570\u683c\u5f0f\u9519\u8bef")
		return
	}

	updates := map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Match != nil {
		updates["match"] = *req.Match
	}
	if req.Action != nil {
		updates["action"] = *req.Action
	}
	if req.ActionValue != nil {
		updates["action_value"] = *req.ActionValue
	}
	if req.MatchJSON != nil {
		updates["match_json"] = *req.MatchJSON
	}
	if req.ActionJSON != nil {
		updates["action_json"] = *req.ActionJSON
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		BadRequest(c, "\u6ca1\u6709\u9700\u8981\u66f4\u65b0\u7684\u5b57\u6bb5")
		return
	}

	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&rule).Updates(updates).Error; err != nil {
		InternalError(c, "\u66f4\u65b0\u8def\u7531\u89c4\u5219\u5931\u8d25")
		return
	}

	database.DB.First(&rule, id)
	Success(c, rule)
}

// AdminDeleteRoute deletes a route rule.
// DELETE /api/v1/admin/routes/:id
func AdminDeleteRoute(c *gin.Context) {
	id := c.Param("id")

	var rule model.RouteRule
	if err := database.DB.First(&rule, id).Error; err != nil {
		NotFound(c, "\u8def\u7531\u89c4\u5219\u4e0d\u5b58\u5728")
		return
	}

	if err := database.DB.Delete(&rule).Error; err != nil {
		InternalError(c, "\u5220\u9664\u8def\u7531\u89c4\u5219\u5931\u8d25")
		return
	}

	Success(c, gin.H{"message": "\u8def\u7531\u89c4\u5219\u5df2\u5220\u9664"})
}
