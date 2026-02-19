package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters per IP address
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter returns the rate limiter for the given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

// CleanupOldLimiters removes limiters that haven't been used recently
func (i *IPRateLimiter) CleanupOldLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			i.mu.Lock()
			// Simple cleanup: clear all limiters periodically
			// In production, you might want more sophisticated cleanup
			i.limiters = make(map[string]*rate.Limiter)
			i.mu.Unlock()
		}
	}()
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(
		rate.Limit(float64(requestsPerMinute)/60.0), // Convert to per-second
		requestsPerMinute,
	)
	limiter.CleanupOldLimiters()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.Header("X-RateLimit-Limit", string(rune(requestsPerMinute)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ExpensiveEndpointRateLimit creates a stricter rate limit for expensive operations
func ExpensiveEndpointRateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(
		rate.Limit(float64(requestsPerMinute)/60.0),
		requestsPerMinute,
	)
	limiter.CleanupOldLimiters()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.Header("X-RateLimit-Limit", string(rune(requestsPerMinute)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded for this endpoint. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
