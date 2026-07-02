package admin

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const maxAdminRequestBodyBytes = 1 << 20

func securityMiddleware() gin.HandlerFunc {
	allowedOrigins := parseCSVEnv("ADMIN_CORS_ALLOWED_ORIGINS")
	if len(allowedOrigins) == 0 {
		allowedOrigins = parseCSVEnv("CORS_ALLOWED_ORIGINS")
	}
	allowAnyLocalhost := len(allowedOrigins) == 0
	limiter := newIPRateLimiter(60, time.Minute)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (allowedOrigins[origin] || (allowAnyLocalhost && isLocalOrigin(origin))) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}

		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAdminRequestBodyBytes)
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		if !limiter.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}

func parseCSVEnv(key string) map[string]bool {
	values := map[string]bool{}
	for _, value := range strings.Split(os.Getenv(key), ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			values[value] = true
		}
	}
	return values
}

func isLocalOrigin(origin string) bool {
	return strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:") ||
		strings.HasPrefix(origin, "http://[::1]:")
}

type ipRateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	hits   map[string]rateBucket
}

type rateBucket struct {
	count     int
	resetTime time.Time
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	return &ipRateLimiter{
		limit:  limit,
		window: window,
		hits:   map[string]rateBucket{},
	}
}

func (l *ipRateLimiter) allow(ip string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket := l.hits[ip]
	if now.After(bucket.resetTime) {
		bucket = rateBucket{resetTime: now.Add(l.window)}
	}

	bucket.count++
	l.hits[ip] = bucket
	return bucket.count <= l.limit
}
