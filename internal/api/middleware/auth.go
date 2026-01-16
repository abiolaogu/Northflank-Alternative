// Package middleware provides HTTP middleware for the API.
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/northstack/platform/internal/config"
	"github.com/northstack/platform/internal/domain"
	"github.com/northstack/platform/pkg/errors"
	"github.com/northstack/platform/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	config   *config.AuthConfig
	userRepo domain.UserRepository
	logger   *logger.Logger
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(cfg *config.AuthConfig, userRepo domain.UserRepository, log *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config:   cfg,
		userRepo: userRepo,
		logger:   log,
	}
}

// RequireAuth returns a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errors.CodeUnauthorized,
				"message": "authentication required",
			})
			return
		}

		// Check if it's an API key
		if strings.HasPrefix(token, "op_") {
			m.validateAPIKey(c, token)
			return
		}

		// Otherwise, treat as JWT
		m.validateJWT(c, token)
	}
}

// RequireRole returns a middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    errors.CodeForbidden,
				"message": "access denied",
			})
			return
		}

		role := userRole.(domain.UserRole)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":    errors.CodeForbidden,
			"message": "insufficient permissions",
		})
	}
}

// validateJWT validates a JWT token
func (m *AuthMiddleware) validateJWT(c *gin.Context, token string) {
	// In a full implementation, this would:
	// 1. Parse the JWT token
	// 2. Validate the signature using the secret
	// 3. Check expiration
	// 4. Extract user claims

	// Simplified implementation for demonstration
	// In production, use a proper JWT library like github.com/golang-jwt/jwt/v5

	// For now, we'll do a basic validation and look up the user
	// This is a placeholder - implement proper JWT validation
	userID, err := uuid.Parse(token) // This is wrong - just for demonstration
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    errors.CodeUnauthorized,
			"message": "invalid token",
		})
		return
	}

	user, err := m.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    errors.CodeUnauthorized,
			"message": "user not found",
		})
		return
	}

	if !user.IsActive {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    errors.CodeUnauthorized,
			"message": "user account is disabled",
		})
		return
	}

	// Set user context
	c.Set("user_id", user.ID)
	c.Set("user_email", user.Email)
	c.Set("user_role", user.Role)

	c.Next()
}

// validateAPIKey validates an API key
func (m *AuthMiddleware) validateAPIKey(c *gin.Context, apiKey string) {
	if !m.config.APIKeyEnabled {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    errors.CodeUnauthorized,
			"message": "API key authentication is disabled",
		})
		return
	}

	// In a full implementation, this would:
	// 1. Look up the API key in the database
	// 2. Check if it's valid and not expired
	// 3. Get the associated user

	// For demonstration, we'll accept any key starting with "op_"
	// In production, implement proper API key validation

	// Extract user ID from key (placeholder implementation)
	keyParts := strings.Split(apiKey, "_")
	if len(keyParts) < 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    errors.CodeUnauthorized,
			"message": "invalid API key format",
		})
		return
	}

	m.logger.Debug().Str("api_key_prefix", keyParts[0]).Msg("API key authentication attempt")

	// Set a default user for API key authentication
	// In production, look up the actual user from the API key
	c.Set("user_id", uuid.New())
	c.Set("user_role", domain.UserRoleMember)
	c.Set("auth_method", "api_key")

	c.Next()
}

// extractToken extracts the authentication token from the request
func extractToken(c *gin.Context) string {
	// Check Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check X-API-Key header
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// Check query parameter (for convenience in some cases)
	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	config *config.AuthConfig
	logger *logger.Logger
	// In production, use Redis or similar for distributed rate limiting
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware
func NewRateLimitMiddleware(cfg *config.AuthConfig, log *logger.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		config: cfg,
		logger: log,
	}
}

// RateLimit returns a middleware that enforces rate limiting
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	if !m.config.RateLimitEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Simple in-memory rate limiter
	// In production, use a distributed rate limiter
	type clientInfo struct {
		requests int
		window   time.Time
	}
	clients := make(map[string]*clientInfo)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		info, exists := clients[clientIP]
		if !exists || time.Since(info.window) > m.config.RateLimitWindow {
			clients[clientIP] = &clientInfo{
				requests: 1,
				window:   time.Now(),
			}
			c.Next()
			return
		}

		info.requests++
		if info.requests > m.config.RateLimitRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    errors.CodeRateLimited,
				"message": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Int("status", status).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP request")
	}
}
