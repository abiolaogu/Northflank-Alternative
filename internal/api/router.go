// Package api provides the HTTP API server for the Platform Orchestrator.
package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/northstack/platform/internal/api/handlers"
	"github.com/northstack/platform/internal/api/middleware"
	"github.com/northstack/platform/internal/config"
	"github.com/northstack/platform/internal/domain"
	"github.com/northstack/platform/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests in seconds",
		},
		[]string{"path", "method"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(httpDuration)
}

// Router holds all the dependencies for the API router
type Router struct {
	config      *config.Config
	logger      *logger.Logger
	projectRepo domain.ProjectRepository
	serviceRepo domain.ServiceRepository
	userRepo    domain.UserRepository
	eventBus    domain.EventBus
	ciAdapter   domain.CIAdapter
}

// NewRouter creates a new Router
func NewRouter(
	cfg *config.Config,
	log *logger.Logger,
	projectRepo domain.ProjectRepository,
	serviceRepo domain.ServiceRepository,
	userRepo domain.UserRepository,
	eventBus domain.EventBus,
	ciAdapter domain.CIAdapter,
) *Router {
	return &Router{
		config:      cfg,
		logger:      log,
		projectRepo: projectRepo,
		serviceRepo: serviceRepo,
		userRepo:    userRepo,
		eventBus:    eventBus,
		ciAdapter:   ciAdapter,
	}
}

// Setup configures and returns the Gin router
func (r *Router) Setup() *gin.Engine {
	if r.config.Observability.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	// Add logging middleware
	if r.config.Observability.Logging.Level != "" {
		router.Use(middleware.LoggingMiddleware(r.logger))
	}

	if r.config.Server.CORSEnabled {
		router.Use(middleware.CORSMiddleware(r.config.Server.CORSOrigins))
	}

	// Rate limiting
	rateLimiter := middleware.NewRateLimitMiddleware(&r.config.Auth, r.logger)
	router.Use(rateLimiter.RateLimit())

	// Health checks (no auth required)
	healthHandler := handlers.NewHealthHandler("1.0.0", "production")
	router.GET("/health", healthHandler.Live)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

	// Metrics endpoint
	if r.config.Observability.Metrics.Enabled {
		// Expose Prometheus metrics at the configured path
		router.GET(r.config.Observability.Metrics.Path, gin.WrapH(promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		)))
		// Add middleware to record metrics for all requests
		router.Use(func(c *gin.Context) {
			start := time.Now()
			c.Next() // Process request

			path := c.FullPath()
			if path == "" {
				path = "unknown" // Fallback for routes without a full path (e.g., 404s)
			}

			duration := time.Since(start).Seconds()
			status := fmt.Sprintf("%d", c.Writer.Status())

			httpRequests.WithLabelValues(path, c.Request.Method, status).Inc()
			httpDuration.WithLabelValues(path, c.Request.Method).Observe(duration)
		})
	}

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(&r.config.Auth, r.userRepo, r.logger)

	// Auth handler (public routes)
	authHandler := handlers.NewAuthHandler(r.userRepo, &r.config.Auth, r.logger)
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/refresh", authHandler.RefreshToken)
	v1.POST("/webhooks/:source", r.handleWebhook)

	// GitHub webhook handler
	githubWebhook := handlers.NewGitHubWebhookHandler(r.config.Integrations.Coolify.WebhookSecret, r.logger)
	v1.POST("/webhooks/github", githubWebhook.HandleWebhook)

	// Protected routes
	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Projects
		projectHandler := handlers.NewProjectHandler(r.projectRepo, r.eventBus, r.logger)
		protected.POST("/projects", projectHandler.Create)
		protected.GET("/projects", projectHandler.List)
		protected.GET("/projects/:id", projectHandler.Get)
		protected.GET("/projects/slug/:slug", projectHandler.GetBySlug)
		protected.PATCH("/projects/:id", projectHandler.Update)
		protected.DELETE("/projects/:id", projectHandler.Delete)

		// Services
		serviceHandler := handlers.NewServiceHandler(r.serviceRepo, r.projectRepo, r.ciAdapter, r.eventBus, r.logger)
		protected.POST("/projects/:project_id/services", serviceHandler.Create)
		protected.GET("/projects/:project_id/services", serviceHandler.ListByProject)
		protected.GET("/services/:id", serviceHandler.Get)
		protected.PATCH("/services/:id", serviceHandler.Update)
		protected.DELETE("/services/:id", serviceHandler.Delete)
		protected.POST("/services/:id/builds", serviceHandler.TriggerBuild)
		protected.POST("/services/:id/scale", serviceHandler.Scale)

		// User management
		protected.GET("/users/me", authHandler.GetCurrentUser)
		protected.PATCH("/users/me", authHandler.UpdateCurrentUser)
		protected.POST("/auth/logout", authHandler.Logout)

		// Clusters (admin only)
		adminOnly := protected.Group("")
		adminOnly.Use(authMiddleware.RequireRole(domain.UserRoleAdmin))
		{
			adminOnly.POST("/clusters", r.handleCreateCluster)
			adminOnly.GET("/clusters", r.handleListClusters)
			adminOnly.GET("/clusters/:id", r.handleGetCluster)
			adminOnly.DELETE("/clusters/:id", r.handleDeleteCluster)
			adminOnly.GET("/clusters/:id/kubeconfig", r.handleGetClusterKubeconfig)

			// Database management
			adminOnly.POST("/projects/:project_id/databases", r.handleCreateDatabase)
			adminOnly.GET("/projects/:project_id/databases", r.handleListDatabases)
			adminOnly.GET("/databases/:id", r.handleGetDatabase)
			adminOnly.DELETE("/databases/:id", r.handleDeleteDatabase)
			adminOnly.POST("/databases/:id/scale", r.handleScaleDatabase)
		}
	}

	return router
}

// Placeholder handlers for cluster and database - will be injected via DI
func (r *Router) handleWebhook(c *gin.Context) { c.JSON(200, gin.H{"message": "webhook received"}) }
func (r *Router) handleCreateCluster(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject ClusterHandler"})
}
func (r *Router) handleListClusters(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject ClusterHandler"})
}
func (r *Router) handleGetCluster(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject ClusterHandler"})
}
func (r *Router) handleDeleteCluster(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject ClusterHandler"})
}
func (r *Router) handleGetClusterKubeconfig(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented"})
}
func (r *Router) handleCreateDatabase(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject DatabaseHandler"})
}
func (r *Router) handleListDatabases(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject DatabaseHandler"})
}
func (r *Router) handleGetDatabase(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject DatabaseHandler"})
}
func (r *Router) handleDeleteDatabase(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject DatabaseHandler"})
}
func (r *Router) handleScaleDatabase(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented - inject DatabaseHandler"})
}
