// Package middleware provides HTTP middleware for the NorthStack API
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerSecond int
	BurstSize         int
	CleanupInterval   time.Duration
}

// DefaultRateLimiterConfig returns default configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 100,
		BurstSize:         200,
		CleanupInterval:   time.Minute * 5,
	}
}

// clientLimiter tracks per-client rate limiters
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter provides rate limiting middleware
type RateLimiter struct {
	config   RateLimiterConfig
	clients  map[string]*clientLimiter
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		config:   config,
		clients:  make(map[string]*clientLimiter),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop removes stale client entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanup removes clients not seen recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-rl.config.CleanupInterval * 2)
	for key, client := range rl.clients {
		if client.lastSeen.Before(threshold) {
			delete(rl.clients, key)
		}
	}
}

// getLimiter returns the rate limiter for a client
func (rl *RateLimiter) getLimiter(clientID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if client, exists := rl.clients[clientID]; exists {
		client.lastSeen = time.Now()
		return client.limiter
	}

	limiter := rate.NewLimiter(
		rate.Limit(rl.config.RequestsPerSecond),
		rl.config.BurstSize,
	)

	rl.clients[clientID] = &clientLimiter{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use IP + User-Agent as client identifier
		clientID := c.ClientIP() + "-" + c.GetHeader("User-Agent")

		limiter := rl.getLimiter(clientID)

		if !limiter.Allow() {
			c.Header("X-RateLimit-Limit", string(rune(rl.config.RequestsPerSecond)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please slow down.",
			})
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", string(rune(rl.config.RequestsPerSecond)))
		c.Next()
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// IPRateLimiter provides simple IP-based rate limiting
func IPRateLimiter(requestsPerSecond int) gin.HandlerFunc {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         requestsPerSecond * 2,
		CleanupInterval:   time.Minute * 5,
	})
	return rl.Middleware()
}
