package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/northstack/platform/internal/domain"
	"github.com/northstack/platform/pkg/logger"
	"github.com/northstack/platform/pkg/yugabytedb"
)

// DatabaseHandler handles database management endpoints
type DatabaseHandler struct {
	dbService *yugabytedb.DatabaseService
	eventBus  domain.EventBus
	logger    *logger.Logger
}

// NewDatabaseHandler creates a new DatabaseHandler
func NewDatabaseHandler(dbService *yugabytedb.DatabaseService, eventBus domain.EventBus, log *logger.Logger) *DatabaseHandler {
	return &DatabaseHandler{
		dbService: dbService,
		eventBus:  eventBus,
		logger:    log,
	}
}

// CreateDatabaseRequest represents a database creation request
type CreateDatabaseRequest struct {
	Name             string `json:"name" binding:"required"`
	Size             string `json:"size" binding:"required,oneof=small medium large xlarge"`
	StorageGB        int    `json:"storage_gb" binding:"required,min=10,max=1000"`
	HighAvailability bool   `json:"high_availability"`
	BackupEnabled    bool   `json:"backup_enabled"`
	TLSEnabled       bool   `json:"tls_enabled"`
	Version          string `json:"version"`
}

// CreateDatabase creates a new YugabyteDB cluster
func (h *DatabaseHandler) CreateDatabase(c *gin.Context) {
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
		return
	}

	var req CreateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get team ID from context
	teamID := ""
	if tid, exists := c.Get("team_id"); exists {
		teamID = tid.(string)
	}

	input := &yugabytedb.CreateDatabaseInput{
		Name:             req.Name,
		ProjectID:        projectID,
		TeamID:           teamID,
		Size:             req.Size,
		StorageGB:        req.StorageGB,
		HighAvailability: req.HighAvailability,
		BackupEnabled:    req.BackupEnabled,
		TLSEnabled:       req.TLSEnabled,
		Version:          req.Version,
	}

	db, err := h.dbService.CreateDatabase(c.Request.Context(), input)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create database"})
		return
	}

	// Publish event
	h.publishEvent(c.Request.Context(), "database.created", map[string]interface{}{
		"database_id": db.ID,
		"project_id":  projectID,
		"name":        req.Name,
	})

	c.JSON(http.StatusCreated, db)
}

// ListDatabases lists all databases for a project
func (h *DatabaseHandler) ListDatabases(c *gin.Context) {
	projectID := c.Param("project_id")

	databases, err := h.dbService.ListDatabases(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list databases")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list databases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"databases": databases,
		"total":     len(databases),
	})
}

// GetDatabase retrieves a specific database
func (h *DatabaseHandler) GetDatabase(c *gin.Context) {
	databaseID := c.Param("id")
	if databaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database ID required"})
		return
	}

	db, err := h.dbService.GetDatabase(c.Request.Context(), databaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
		return
	}

	c.JSON(http.StatusOK, db)
}

// DeleteDatabase deletes a database
func (h *DatabaseHandler) DeleteDatabase(c *gin.Context) {
	databaseID := c.Param("id")
	if databaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database ID required"})
		return
	}

	if err := h.dbService.DeleteDatabase(c.Request.Context(), databaseID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete database"})
		return
	}

	// Publish event
	h.publishEvent(c.Request.Context(), "database.deleted", map[string]interface{}{
		"database_id": databaseID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Database deleted"})
}

// ScaleDatabaseRequest represents a scaling request
type ScaleDatabaseRequest struct {
	Replicas int `json:"replicas" binding:"required,min=1,max=10"`
}

// ScaleDatabase scales a database cluster
func (h *DatabaseHandler) ScaleDatabase(c *gin.Context) {
	databaseID := c.Param("id")
	if databaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database ID required"})
		return
	}

	var req ScaleDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.dbService.ScaleDatabase(c.Request.Context(), databaseID, req.Replicas); err != nil {
		h.logger.Error().Err(err).Msg("Failed to scale database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scale database"})
		return
	}

	// Publish event
	h.publishEvent(c.Request.Context(), "database.scaled", map[string]interface{}{
		"database_id": databaseID,
		"replicas":    req.Replicas,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":  "Database scaling initiated",
		"replicas": req.Replicas,
	})
}

// GetConnectionInfo returns connection details for a database
func (h *DatabaseHandler) GetConnectionInfo(c *gin.Context) {
	databaseID := c.Param("id")
	if databaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database ID required"})
		return
	}

	db, err := h.dbService.GetDatabase(c.Request.Context(), databaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ysql_endpoint":     db.YSQLEndpoint,
		"ycql_endpoint":     db.YCQLEndpoint,
		"port":              db.Port,
		"database":          db.Database,
		"secret_name":       db.SecretName,
		"connection_string": "View secret for full connection string",
	})
}

func (h *DatabaseHandler) publishEvent(ctx context.Context, eventType string, data map[string]interface{}) {
	event := &domain.Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	if err := h.eventBus.Publish(ctx, eventType, event); err != nil {
		h.logger.Warn().Err(err).Str("event", eventType).Msg("Failed to publish event")
	}
}
