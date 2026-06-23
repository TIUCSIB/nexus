package handler

import (
	"nexus/internal/config"
	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/pkg/crypto"
	"nexus/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请输入邮箱和密码")
		return
	}

	var user model.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		Unauthorized(c, "邮箱或密码错误")
		return
	}

	if user.Status != 1 {
		Forbidden(c, "账号已被禁用")
		return
	}

	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		Unauthorized(c, "邮箱或密码错误")
		return
	}

	expireHours := config.Global.JWT.ExpireHours
	if expireHours <= 0 {
		expireHours = 72
	}

	accessToken, err := jwt.Generate(user.ID, user.IsAdmin, expireHours)
	if err != nil {
		InternalError(c, "生成令牌失败")
		return
	}

	refreshToken, err := jwt.Generate(user.ID, user.IsAdmin, expireHours*7)
	if err != nil {
		InternalError(c, "生成刷新令牌失败")
		return
	}

	Success(c, loginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请输入刷新令牌")
		return
	}

	claims, err := jwt.Parse(req.RefreshToken)
	if err != nil {
		Unauthorized(c, "刷新令牌已过期或无效，请重新登录")
		return
	}

	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		Unauthorized(c, "用户不存在")
		return
	}

	if user.Status != 1 {
		Forbidden(c, "账号已被禁用")
		return
	}

	expireHours := config.Global.JWT.ExpireHours
	if expireHours <= 0 {
		expireHours = 72
	}

	newAccessToken, err := jwt.Generate(user.ID, user.IsAdmin, expireHours)
	if err != nil {
		InternalError(c, "生成新令牌失败")
		return
	}

	newRefreshToken, err := jwt.Generate(user.ID, user.IsAdmin, expireHours*7)
	if err != nil {
		InternalError(c, "生成新刷新令牌失败")
		return
	}

	Success(c, loginResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		User:         user,
	})
}
