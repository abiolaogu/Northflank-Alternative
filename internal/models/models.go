package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Project represents a logical grouping of applications
type Project struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	TeamID      string         `json:"team_id" gorm:"index"`
	Settings    datatypes.JSON `json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Application represents a deployable service
type Application struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	ProjectID   string         `json:"project_id" gorm:"index"`
	Name        string         `json:"name"`
	Repository  string         `json:"repository"`
	Branch      string         `json:"branch"`
	BuildPack   string         `json:"build_pack"`
	Environment map[string]string `json:"environment" gorm:"serializer:json"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Deployment represents a specific version deployment of an application
type Deployment struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	ApplicationID string    `json:"application_id" gorm:"index"`
	Revision      string    `json:"revision"`
	Status        string    `json:"status"` // pending, building, deploying, running, failed
	Strategy      string    `json:"strategy"` // rolling, recreate, blue-green
	StartedAt     time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	Logs          string    `json:"-"` // Stored separately or in object storage
}

// Cluster represents a Kubernetes cluster managed by Rancher/RKE2
type Cluster struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Provider  string    `json:"provider"` // rke2, eks, gke, aks, unknown
	Version   string    `json:"version"`
	Status    string    `json:"status"` // provisioning, active, updating, error
	NodeCount int       `json:"node_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Team represents a group of users
type Team struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Members   []User    `json:"members" gorm:"many2many:team_members;"`
}

// User represents a platform user
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	Name      string    `json:"name"`
	Roles     []string  `json:"roles" gorm:"serializer:json"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuditLog represents a system action for compliance
type AuditLog struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	ActorID   string         `json:"actor_id" gorm:"index"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource"`
	ResourceID string        `json:"resource_id"`
	Details   datatypes.JSON `json:"details"`
	Timestamp time.Time      `json:"timestamp" gorm:"index"`
	IPAddress string         `json:"ip_address"`
}

// BeforeCreate hooks to generate IDs if missing
func (base *Project) BeforeCreate() error {
	if base.ID == "" {
		base.ID = uuid.New().String()
	}
	return nil
}

func (base *Application) BeforeCreate() error {
	if base.ID == "" {
		base.ID = uuid.New().String()
	}
	return nil
}
