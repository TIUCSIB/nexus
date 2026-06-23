package handler

import (
	"strconv"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/pkg/crypto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func AdminListUsers(c *gin.Context) {
	page, pageSize := parsePagination(c)
	q := c.Query("q")

	query := database.DB.Model(&model.User{})
	if q != "" {
		query = query.Where("email LIKE ?", "%"+q+"%")
	}

	var total int64
	query.Count(&total)

	var users []model.User
	offset := (page - 1) * pageSize
	query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users)

	SuccessPage(c, users, total, page, pageSize)
}

func AdminGetUser(c *gin.Context) {
	id := c.Param("id")

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		NotFound(c, "\u7528\u6237\u4e0d\u5b58\u5728")
		return
	}

	Success(c, user)
}

type createUserRequest struct {
	Email          string `json:"email" binding:"required"`
	Password       string `json:"password" binding:"required"`
	PlanID         *uint  `json:"plan_id"`
	TrafficLimit   int64  `json:"traffic_limit"`
	ExpiredAt      string `json:"expired_at"`
	IsAdmin        bool   `json:"is_admin"`
	Status         *int   `json:"status"`
	DeviceLimit    *int   `json:"device_limit"`
	SpeedLimitUp   *int   `json:"speed_limit_up"`
	SpeedLimitDown *int   `json:"speed_limit_down"`
}

func AdminCreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "\u8bf7\u8f93\u5165\u90ae\u7bb1\u548c\u5bc6\u7801")
		return
	}

	if len(req.Password) < 8 {
		BadRequest(c, "\u5bc6\u7801\u957f\u5ea6\u4e0d\u80fd\u5c11\u4e8e8\u4f4d")
		return
	}

	var count int64
	database.DB.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		BadRequest(c, "\u8be5\u90ae\u7bb1\u5df2\u88ab\u6ce8\u518c")
		return
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		InternalError(c, "\u5bc6\u7801\u52a0\u5bc6\u5931\u8d25")
		return
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	deviceLimit := 0
	if req.DeviceLimit != nil {
		deviceLimit = *req.DeviceLimit
	}

	speedLimitUp := 0
	if req.SpeedLimitUp != nil {
		speedLimitUp = *req.SpeedLimitUp
	}

	speedLimitDown := 0
	if req.SpeedLimitDown != nil {
		speedLimitDown = *req.SpeedLimitDown
	}

	user := model.User{
		UUID:           uuid.New().String(),
		Email:          req.Email,
		PasswordHash:   hash,
		Balance:        0,
		PlanID:         req.PlanID,
		TrafficLimit:   req.TrafficLimit,
		IsAdmin:        req.IsAdmin,
		Token:          uuid.New().String(),
		Status:         status,
		DeviceLimit:    deviceLimit,
		SpeedLimitUp:   speedLimitUp,
		SpeedLimitDown: speedLimitDown,
	}

	if req.ExpiredAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiredAt)
		if err != nil {
			BadRequest(c, "\u8fc7\u671f\u65f6\u95f4\u683c\u5f0f\u9519\u8bef\uff0c\u8bf7\u4f7f\u7528 YYYY-MM-DD")
			return
		}
		user.ExpiredAt = &t
	}

	if err := database.DB.Create(&user).Error; err != nil {
		InternalError(c, "\u521b\u5efa\u7528\u6237\u5931\u8d25")
		return
	}

	Success(c, user)
}

type updateUserRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	PlanID         *uint  `json:"plan_id"`
	TrafficLimit   *int64 `json:"traffic_limit"`
	TrafficUsed    *int64 `json:"traffic_used"`
	ExpiredAt      string `json:"expired_at"`
	IsAdmin        *bool  `json:"is_admin"`
	Status         *int   `json:"status"`
	Balance        *int64 `json:"balance"`
	DeviceLimit    *int   `json:"device_limit"`
	SpeedLimitUp   *int   `json:"speed_limit_up"`
	SpeedLimitDown *int   `json:"speed_limit_down"`
}

func AdminUpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		NotFound(c, "\u7528\u6237\u4e0d\u5b58\u5728")
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "\u8bf7\u6c42\u53c2\u6570\u683c\u5f0f\u9519\u8bef")
		return
	}

	updates := map[string]interface{}{}

	if req.Email != "" && req.Email != user.Email {
		var count int64
		database.DB.Model(&model.User{}).Where("email = ? AND id != ?", req.Email, user.ID).Count(&count)
		if count > 0 {
			BadRequest(c, "\u8be5\u90ae\u7bb1\u5df2\u88ab\u5176\u4ed6\u8d26\u53f7\u4f7f\u7528")
			return
		}
		updates["email"] = req.Email
	}

	if req.Password != "" {
		if len(req.Password) < 8 {
			BadRequest(c, "\u5bc6\u7801\u957f\u5ea6\u4e0d\u80fd\u5c11\u4e8e8\u4f4d")
			return
		}
		hash, err := crypto.HashPassword(req.Password)
		if err != nil {
			InternalError(c, "\u5bc6\u7801\u52a0\u5bc6\u5931\u8d25")
			return
		}
		updates["password_hash"] = hash
	}

	if req.PlanID != nil {
		updates["plan_id"] = *req.PlanID
	}
	if req.TrafficLimit != nil {
		updates["traffic_limit"] = *req.TrafficLimit
	}
	if req.TrafficUsed != nil {
		updates["traffic_used"] = *req.TrafficUsed
	}
	if req.Balance != nil {
		updates["balance"] = *req.Balance
	}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.DeviceLimit != nil {
		updates["device_limit"] = *req.DeviceLimit
	}
	if req.SpeedLimitUp != nil {
		updates["speed_limit_up"] = *req.SpeedLimitUp
	}
	if req.SpeedLimitDown != nil {
		updates["speed_limit_down"] = *req.SpeedLimitDown
	}
	if req.ExpiredAt != "" {
		if req.ExpiredAt == "null" || req.ExpiredAt == "" {
			updates["expired_at"] = nil
		} else {
			t, err := time.Parse("2006-01-02", req.ExpiredAt)
			if err != nil {
				BadRequest(c, "\u8fc7\u671f\u65f6\u95f4\u683c\u5f0f\u9519\u8bef\uff0c\u8bf7\u4f7f\u7528 YYYY-MM-DD")
				return
			}
			updates["expired_at"] = t
		}
	}

	if len(updates) == 0 {
		BadRequest(c, "\u6ca1\u6709\u9700\u8981\u66f4\u65b0\u7684\u5b57\u6bb5")
		return
	}

	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		InternalError(c, "\u66f4\u65b0\u7528\u6237\u5931\u8d25")
		return
	}

	database.DB.First(&user, id)
	Success(c, user)
}

func AdminDeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		NotFound(c, "\u7528\u6237\u4e0d\u5b58\u5728")
		return
	}

	adminID := c.GetUint("user_id")
	if user.ID == adminID {
		BadRequest(c, "\u4e0d\u80fd\u5220\u9664\u81ea\u5df1\u7684\u8d26\u53f7")
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		InternalError(c, "\u5220\u9664\u7528\u6237\u5931\u8d25")
		return
	}

	Success(c, gin.H{"message": "\u7528\u6237\u5df2\u5220\u9664"})
}