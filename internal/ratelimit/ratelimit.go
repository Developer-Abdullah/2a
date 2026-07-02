// Package ratelimit provides a small fixed-window rate limiter and a Gin middleware for it. It is
// used to put a tight brute-force cap on sensitive endpoints (admin login, activation-code
// validation) on top of the coarse per-IP limit applied to every request.
//
// Tradeoff: the limiter is in-memory and therefore per-process. Behind multiple replicas the
// effective limit is (limit x replicas). For the roadmap's scale that is acceptable for a
// brute-force speed bump; a Redis-backed counter (INCR + EXPIRE) is the drop-in upgrade when the
// limit must hold cluster-wide, and Allow/keying are structured so that swap is localized here.
package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	count int
	reset time.Time
}

type Limiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	hits   map[string]bucket
}

func New(limit int, window time.Duration) *Limiter {
	return &Limiter{limit: limit, window: window, hits: map[string]bucket{}}
}

// Allow reports whether the key is under its limit for the current window and counts the hit.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.hits[key]
	if now.After(b.reset) {
		b = bucket{reset: now.Add(l.window)}
	}
	b.count++
	l.hits[key] = b
	return b.count <= l.limit
}

// Middleware limits requests per client IP. Exceeding the limit aborts with 429 before the handler
// runs, so a failed login/activation never reaches the (relatively expensive) credential check.
func Middleware(limit int, window time.Duration) gin.HandlerFunc {
	l := New(limit, window)
	return func(c *gin.Context) {
		if !l.Allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "error": gin.H{"message": "too many attempts, please try again later"}})
			return
		}
		c.Next()
	}
}
