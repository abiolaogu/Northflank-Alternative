// Package api provides the HTTP API server for the Platform Orchestrator.
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/openpaas/platform-orchestrator/internal/api/handlers"
	"github.com/openpaas/platform-orchestrator/internal/api/middleware"
	"github.com/openpaas/platform-orchestrator/internal/config"
	"github.com/openpaas/platform-orchestrator/internal/domain"
	"github.com/openpaas/platform-orchestrator/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
	if r.config.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware(r.logger))

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
	if r.config.Metrics.Enabled {
		router.GET(r.config.Metrics.Path, gin.WrapH(promhttp.Handler()))
	}

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(&r.config.Auth, r.userRepo, r.logger)

	// Public routes (no auth)
	v1.POST("/auth/login", r.handleLogin)
	v1.POST("/auth/register", r.handleRegister)
	v1.POST("/webhooks/:source", r.handleWebhook)

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

		// Clusters (admin only)
		adminOnly := protected.Group("")
		adminOnly.Use(authMiddleware.RequireRole(domain.UserRoleAdmin))
		{
			adminOnly.POST("/clusters", r.handleCreateCluster)
			adminOnly.GET("/clusters", r.handleListClusters)
			adminOnly.GET("/clusters/:id", r.handleGetCluster)
			adminOnly.DELETE("/clusters/:id", r.handleDeleteCluster)
		}

		// User management
		protected.GET("/users/me", r.handleGetCurrentUser)
		protected.PATCH("/users/me", r.handleUpdateCurrentUser)
	}

	return router
}

// Placeholder handlers - implement fully in production
func (r *Router) handleLogin(c *gin.Context)         { c.JSON(200, gin.H{"message": "login endpoint"}) }
func (r *Router) handleRegister(c *gin.Context)      { c.JSON(200, gin.H{"message": "register endpoint"}) }
func (r *Router) handleWebhook(c *gin.Context)       { c.JSON(200, gin.H{"message": "webhook received"}) }
func (r *Router) handleCreateCluster(c *gin.Context) { c.JSON(200, gin.H{"message": "create cluster"}) }
func (r *Router) handleListClusters(c *gin.Context)  { c.JSON(200, gin.H{"message": "list clusters"}) }
func (r *Router) handleGetCluster(c *gin.Context)    { c.JSON(200, gin.H{"message": "get cluster"}) }
func (r *Router) handleDeleteCluster(c *gin.Context) { c.JSON(200, gin.H{"message": "delete cluster"}) }
func (r *Router) handleGetCurrentUser(c *gin.Context) {
	c.JSON(200, gin.H{"message": "get current user"})
}
func (r *Router) handleUpdateCurrentUser(c *gin.Context) {
	c.JSON(200, gin.H{"message": "update current user"})
}
