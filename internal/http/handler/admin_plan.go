package handler

import (
	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

func AdminListPlans(c *gin.Context) {
	page, pageSize := parsePagination(c)

	var total int64
	database.DB.Model(&model.Plan{}).Count(&total)

	var plans []model.Plan
	offset := (page - 1) * pageSize
	database.DB.Order("sort ASC, id ASC").Offset(offset).Limit(pageSize).Find(&plans)

	SuccessPage(c, plans, total, page, pageSize)
}

type createPlanRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	TrafficLimit int64  `json:"traffic_limit"`
	DurationDays int    `json:"duration_days" binding:"required"`
	Price        int64  `json:"price"`
	Sort         int    `json:"sort"`
	Status       *int   `json:"status"`
}

func AdminCreatePlan(c *gin.Context) {
	var req createPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请输入套餐名称和时长")
		return
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	plan := model.Plan{
		Name:         req.Name,
		Description:  req.Description,
		TrafficLimit: req.TrafficLimit,
		DurationDays: req.DurationDays,
		Price:        req.Price,
		Sort:         req.Sort,
		Status:       status,
	}

	if err := database.DB.Create(&plan).Error; err != nil {
		InternalError(c, "创建套餐失败")
		return
	}

	Success(c, plan)
}

type updatePlanRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	TrafficLimit *int64 `json:"traffic_limit"`
	DurationDays *int   `json:"duration_days"`
	Price        *int64 `json:"price"`
	Sort         *int   `json:"sort"`
	Status       *int   `json:"status"`
}

func AdminUpdatePlan(c *gin.Context) {
	id := c.Param("id")

	var plan model.Plan
	if err := database.DB.First(&plan, id).Error; err != nil {
		NotFound(c, "套餐不存在")
		return
	}

	var req updatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请求参数格式错误")
		return
	}

	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.TrafficLimit != nil {
		updates["traffic_limit"] = *req.TrafficLimit
	}
	if req.DurationDays != nil {
		updates["duration_days"] = *req.DurationDays
	}
	if req.Price != nil {
		updates["price"] = *req.Price
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

	if err := database.DB.Model(&plan).Updates(updates).Error; err != nil {
		InternalError(c, "更新套餐失败")
		return
	}

	database.DB.First(&plan, id)
	Success(c, plan)
}

func AdminDeletePlan(c *gin.Context) {
	id := c.Param("id")

	var plan model.Plan
	if err := database.DB.First(&plan, id).Error; err != nil {
		NotFound(c, "套餐不存在")
		return
	}

	var userCount int64
	database.DB.Model(&model.User{}).Where("plan_id = ?", plan.ID).Count(&userCount)
	if userCount > 0 {
		BadRequest(c, "该套餐下仍有用户，请先迁移用户到其他套餐")
		return
	}

	if err := database.DB.Delete(&plan).Error; err != nil {
		InternalError(c, "删除套餐失败")
		return
	}

	Success(c, gin.H{"message": "套餐已删除"})
}
