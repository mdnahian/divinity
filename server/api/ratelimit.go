package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/divinity/core/config"
)

// RateLimiter provides per-token rate limiting using token-bucket algorithm.
type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func NewRateLimiter(cfg *config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(cfg.RequestsPerSecond),
		burst:    cfg.BurstSize,
	}
}

func (rl *RateLimiter) getLimiter(token string) *rate.Limiter {
	rl.mu.RLock()
	lim, ok := rl.limiters[token]
	rl.mu.RUnlock()
	if ok {
		return lim
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()
	// Double-check after acquiring write lock
	if lim, ok = rl.limiters[token]; ok {
		return lim
	}
	lim = rate.NewLimiter(rl.rps, rl.burst)
	rl.limiters[token] = lim
	return lim
}

// Evict removes the rate limiter for a token (e.g., when an NPC dies).
func (rl *RateLimiter) Evict(token string) {
	rl.mu.Lock()
	delete(rl.limiters, token)
	rl.mu.Unlock()
}

// Middleware returns a gin middleware that enforces rate limits per Bearer token.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, exists := c.Get("token")
		if !exists {
			c.Next()
			return
		}

		lim := rl.getLimiter(token.(string))
		if !lim.Allow() {
			retryAfter := 1.0 / float64(rl.rps)
			c.Header("Retry-After", fmt.Sprintf("%.1f", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
			})
			return
		}
		c.Next()
	}
}
