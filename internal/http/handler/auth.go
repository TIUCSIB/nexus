package handler

import (
	"log"
	"strings"
	"sync"
	"time"

	"nexus/internal/config"
	"nexus/internal/database"
	"nexus/internal/model"
	"nexus/internal/pkg/crypto"
	"nexus/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// account-level brute force protection
type loginAttempt struct {
	count    int
	lockedAt time.Time
}

var (
	loginMu      sync.Mutex
	loginAttempts = make(map[string]*loginAttempt) // email → attempt info
	maxLoginAttempts = 5
	loginLockoutDuration = 15 * time.Minute
	loginAttemptWindow   = 30 * time.Minute // reset counter after this window

	// cleanup goroutine for login attempts
	loginCleanupOnce sync.Once
)

func initLoginCleanup() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			loginMu.Lock()
			now := time.Now()
			for email, att := range loginAttempts {
				// Remove entries that have been locked out for too long or expired
				if att.lockedAt.IsZero() && now.Sub(att.lockedAt.Add(loginLockoutDuration)) > loginAttemptWindow {
					delete(loginAttempts, email)
					continue
				}
				if !att.lockedAt.IsZero() && now.After(att.lockedAt.Add(loginLockoutDuration)) {
					delete(loginAttempts, email)
				}
			}
			loginMu.Unlock()
		}
	}()
}

func checkLoginLockout(email string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()

	att, exists := loginAttempts[email]
	if !exists {
		return false
	}

	// If locked, check if lockout period has passed
	if !att.lockedAt.IsZero() {
		if time.Since(att.lockedAt) < loginLockoutDuration {
			return true // still locked
		}
		// Lockout expired, reset
		delete(loginAttempts, email)
		return false
	}

	return false
}

func recordFailedLogin(email string) {
	loginMu.Lock()
	defer loginMu.Unlock()

	att, exists := loginAttempts[email]
	if !exists {
		att = &loginAttempt{}
		loginAttempts[email] = att
	}

	att.count++

	// Lock the account after max failed attempts
	if att.count >= maxLoginAttempts {
		att.lockedAt = time.Now()
		log.Printf("[安全] 账号 %s 因登录失败 %d 次被锁定 %v", maskEmail(email), maxLoginAttempts, loginLockoutDuration)
	}
}

func resetLoginAttempts(email string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	delete(loginAttempts, email)
}

// maskEmail hides part of email for logging
func maskEmail(email string) string {
	if len(email) <= 3 {
		return "***"
	}
	at := strings.Index(email, "@")
	if at <= 0 {
		return email[:1] + "***" + email[len(email)-1:]
	}
	return email[:1] + "***" + email[at-1:]
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
	AdminPath    string      `json:"admin_path"`
	AuthPath     string      `json:"auth_path"`
	UserPath     string      `json:"user_path"`
	AppName      string      `json:"app_name"`
	AppDesc      string      `json:"app_description"`
	SubPath      string      `json:"sub_path"`
}

func Login(c *gin.Context) {
		loginCleanupOnce.Do(initLoginCleanup)

		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequest(c, "请输入邮箱和密码")
			return
		}

		// Check account-level lockout before proceeding
		if checkLoginLockout(req.Email) {
			log.Printf("[安全] 被锁定的账号尝试登录: %s", maskEmail(req.Email))
			Unauthorized(c, "账号已被临时锁定，请15分钟后再试")
			return
		}

		var user model.User
		if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
			recordFailedLogin(req.Email)
			Unauthorized(c, "邮箱或密码错误")
			return
		}

		if user.Status != 1 {
			Forbidden(c, "账号已被禁用")
			return
		}

		if !crypto.CheckPassword(req.Password, user.PasswordHash) {
			recordFailedLogin(req.Email)
			Unauthorized(c, "邮箱或密码错误")
			return
		}

		// Successful login — reset failed attempts
		resetLoginAttempts(req.Email)

	expireHours := config.Global.JWT.ExpireHours
	if expireHours <= 0 {
		expireHours = 72
	}

	accessToken, err := jwt.Generate(user.ID, user.IsAdmin, user.TokenVersion, expireHours)
		if err != nil {
			InternalError(c, "生成令牌失败")
			return
		}

		refreshToken, err := jwt.Generate(user.ID, user.IsAdmin, user.TokenVersion, expireHours*7)
		if err != nil {
			InternalError(c, "生成刷新令牌失败")
			return
		}

	Success(c, loginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user,
		AdminPath:    database.GetSettingDefault("admin_path", "admin"),
		AuthPath:     database.GetSettingDefault("auth_path", "auth"),
		UserPath:     database.GetSettingDefault("user_path", "user"),
		AppName:      database.GetSettingDefault("app_name", "Nexus"),
		AppDesc:      database.GetSettingDefault("app_description", ""),
		SubPath:      database.GetSettingDefault("sub_path", "s"),
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

		// Token rotation: verify token version matches, then increment
		if claims.TokenVersion != user.TokenVersion {
			Unauthorized(c, "刷新令牌已被使用，请重新登录")
			return
		}

		expireHours := config.Global.JWT.ExpireHours
		if expireHours <= 0 {
			expireHours = 72
		}

		// Increment token version to invalidate old refresh tokens
		database.DB.Model(&user).Update("token_version", user.TokenVersion+1)
		user.TokenVersion++

		newAccessToken, err := jwt.Generate(user.ID, user.IsAdmin, user.TokenVersion, expireHours)
		if err != nil {
			InternalError(c, "生成新令牌失败")
			return
		}

		newRefreshToken, err := jwt.Generate(user.ID, user.IsAdmin, user.TokenVersion, expireHours*7)
		if err != nil {
			InternalError(c, "生成新刷新令牌失败")
			return
		}

	Success(c, loginResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		User:         user,
		AdminPath:    database.GetSettingDefault("admin_path", "admin"),
		AuthPath:     database.GetSettingDefault("auth_path", "auth"),
		UserPath:     database.GetSettingDefault("user_path", "user"),
		AppName:      database.GetSettingDefault("app_name", "Nexus"),
		AppDesc:      database.GetSettingDefault("app_description", ""),
		SubPath:      database.GetSettingDefault("sub_path", "s"),
	})
}