package middleware

import (
	"net/http"
	"strings"

	"nexus/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供有效的认证令牌"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := jwt.Parse(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌已过期或无效"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}
