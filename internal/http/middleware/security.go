package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders returns a Gin middleware that sets common security HTTP headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// HSTS — only when TLS is active
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content-Security-Policy (relaxed for SPA with inline scripts)
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"connect-src 'self' ws: wss:; "+
				"font-src 'self' data:;")

		// Prevent caching of API responses
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", time.Now().Format(time.RFC1123))
		}

		c.Next()
	}
}