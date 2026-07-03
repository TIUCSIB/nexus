package middleware

import (
	"net/http"
	"strings"

	"nexus/internal/database"

	"github.com/gin-gonic/gin"
)

func ForceHTTPS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !database.GetSettingBool("force_https", false) {
			c.Next()
			return
		}

		if c.Request.TLS != nil {
			c.Next()
			return
		}

		if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
			c.Next()
			return
		}

		host := c.Request.Host
		host = strings.Split(host, ":")[0]

		target := "https://" + host + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}

		c.Redirect(http.StatusMovedPermanently, target)
		c.Abort()
	}
}