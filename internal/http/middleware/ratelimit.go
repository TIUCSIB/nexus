package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a sliding window rate limiter per IP.
type RateLimiter struct {
	mu       sync.Mutex
	windows  map[string][]time.Time // IP -> request timestamps
	maxReq   int
	window   time.Duration
	stopChan chan struct{}
}

// NewRateLimiter creates a rate limiter that allows maxReq requests per window duration.
func NewRateLimiter(maxReq int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		windows:  make(map[string][]time.Time),
		maxReq:   maxReq,
		window:   window,
		stopChan: make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Stop stops the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// Allow checks if a request from the given key (IP) is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove expired timestamps
	timestamps := rl.windows[key]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.maxReq {
		rl.windows[key] = valid
		return false
	}

	rl.windows[key] = append(valid, now)
	return true
}

// cleanup periodically removes stale entries.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-2 * rl.window)
			for ip, timestamps := range rl.windows {
				var valid []time.Time
				for _, t := range timestamps {
					if t.After(cutoff) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(rl.windows, ip)
				} else {
					rl.windows[ip] = valid
				}
			}
			rl.mu.Unlock()
		case <-rl.stopChan:
			return
		}
	}
}

// RateLimit returns a Gin middleware that rate-limits requests per client IP.
func RateLimit(maxReq int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(maxReq, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    -1,
				"message": "请求过于频繁，请稍后重试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Global rate limiter instances for different tiers
var (
	AuthLimiter   = NewRateLimiter(5, time.Minute)   // 5 req/min for auth
	UserLimiter   = NewRateLimiter(30, time.Minute)   // 30 req/min for user endpoints
	AdminLimiter  = NewRateLimiter(60, time.Minute)   // 60 req/min for admin endpoints
)