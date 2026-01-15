package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/openpaas/platform-orchestrator/internal/domain"
	"github.com/openpaas/platform-orchestrator/pkg/errors"
)

// ServiceRepository implements domain.ServiceRepository using PostgreSQL
type ServiceRepository struct {
	db *PostgresDB
}

// NewServiceRepository creates a new ServiceRepository
func NewServiceRepository(db *PostgresDB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// Create creates a new service
func (r *ServiceRepository) Create(ctx context.Context, service *domain.Service) error {
	buildSource, _ := json.Marshal(service.BuildSource)
	resources, _ := json.Marshal(service.Resources)
	scaling, _ := json.Marshal(service.Scaling)
	healthCheck, _ := json.Marshal(service.HealthCheck)
	envVars, _ := json.Marshal(service.EnvVars)
	secretRefs, _ := json.Marshal(service.SecretRefs)
	ports, _ := json.Marshal(service.Ports)
	labels, _ := json.Marshal(service.Labels)
	annotations, _ := json.Marshal(service.Annotations)
	metadata, _ := json.Marshal(service.Metadata)

	query := `
		INSERT INTO services (
			id, project_id, name, slug, type, status, build_source, resources, scaling,
			health_check, env_vars, secret_refs, ports, labels, annotations, metadata,
			current_build_id, current_version, target_cluster_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`

	_, err := r.db.pool.Exec(ctx, query,
		service.ID,
		service.ProjectID,
		service.Name,
		service.Slug,
		service.Type,
		service.Status,
		buildSource,
		resources,
		scaling,
		healthCheck,
		envVars,
		secretRefs,
		ports,
		labels,
		annotations,
		metadata,
		service.CurrentBuildID,
		service.CurrentVersion,
		service.TargetClusterID,
		service.CreatedAt,
		service.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, "failed to create service")
	}

	return nil
}

// GetByID retrieves a service by ID
func (r *ServiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	query := `
		SELECT id, project_id, name, slug, type, status, build_source, resources, scaling,
			health_check, env_vars, secret_refs, ports, labels, annotations, metadata,
			current_build_id, current_version, target_cluster_id, created_at, updated_at
		FROM services
		WHERE id = $1
	`

	return r.scanService(ctx, query, id)
}

// GetBySlug retrieves a service by project ID and slug
func (r *ServiceRepository) GetBySlug(ctx context.Context, projectID uuid.UUID, slug string) (*domain.Service, error) {
	query := `
		SELECT id, project_id, name, slug, type, status, build_source, resources, scaling,
			health_check, env_vars, secret_refs, ports, labels, annotations, metadata,
			current_build_id, current_version, target_cluster_id, created_at, updated_at
		FROM services
		WHERE project_id = $1 AND slug = $2
	`

	return r.scanService(ctx, query, projectID, slug)
}

func (r *ServiceRepository) scanService(ctx context.Context, query string, args ...interface{}) (*domain.Service, error) {
	service := &domain.Service{}
	var buildSource, resources, scaling, healthCheck, envVars, secretRefs, ports, labels, annotations, metadata []byte

	err := r.db.pool.QueryRow(ctx, query, args...).Scan(
		&service.ID,
		&service.ProjectID,
		&service.Name,
		&service.Slug,
		&service.Type,
		&service.Status,
		&buildSource,
		&resources,
		&scaling,
		&healthCheck,
		&envVars,
		&secretRefs,
		&ports,
		&labels,
		&annotations,
		&metadata,
		&service.CurrentBuildID,
		&service.CurrentVersion,
		&service.TargetClusterID,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("service", fmt.Sprintf("%v", args[0]))
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service")
	}

	json.Unmarshal(buildSource, &service.BuildSource)
	json.Unmarshal(resources, &service.Resources)
	json.Unmarshal(scaling, &service.Scaling)
	json.Unmarshal(healthCheck, &service.HealthCheck)
	json.Unmarshal(envVars, &service.EnvVars)
	json.Unmarshal(secretRefs, &service.SecretRefs)
	json.Unmarshal(ports, &service.Ports)
	json.Unmarshal(labels, &service.Labels)
	json.Unmarshal(annotations, &service.Annotations)
	json.Unmarshal(metadata, &service.Metadata)

	return service, nil
}

// ListByProject retrieves services for a project
func (r *ServiceRepository) ListByProject(ctx context.Context, projectID uuid.UUID, filter domain.ServiceFilter) ([]*domain.Service, error) {
	query := `
		SELECT id, project_id, name, slug, type, status, build_source, resources, scaling,
			health_check, env_vars, secret_refs, ports, labels, annotations, metadata,
			current_build_id, current_version, target_cluster_id, created_at, updated_at
		FROM services
		WHERE project_id = $1
	`
	args := []interface{}{projectID}
	argIndex := 2

	if filter.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR slug ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list services")
	}
	defer rows.Close()

	services := []*domain.Service{}
	for rows.Next() {
		service := &domain.Service{}
		var buildSource, resources, scaling, healthCheck, envVars, secretRefs, ports, labels, annotations, metadata []byte

		err := rows.Scan(
			&service.ID,
			&service.ProjectID,
			&service.Name,
			&service.Slug,
			&service.Type,
			&service.Status,
			&buildSource,
			&resources,
			&scaling,
			&healthCheck,
			&envVars,
			&secretRefs,
			&ports,
			&labels,
			&annotations,
			&metadata,
			&service.CurrentBuildID,
			&service.CurrentVersion,
			&service.TargetClusterID,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan service")
		}

		json.Unmarshal(buildSource, &service.BuildSource)
		json.Unmarshal(resources, &service.Resources)
		json.Unmarshal(scaling, &service.Scaling)
		json.Unmarshal(healthCheck, &service.HealthCheck)
		json.Unmarshal(envVars, &service.EnvVars)
		json.Unmarshal(secretRefs, &service.SecretRefs)
		json.Unmarshal(ports, &service.Ports)
		json.Unmarshal(labels, &service.Labels)
		json.Unmarshal(annotations, &service.Annotations)
		json.Unmarshal(metadata, &service.Metadata)

		services = append(services, service)
	}

	return services, nil
}

// Update updates an existing service
func (r *ServiceRepository) Update(ctx context.Context, service *domain.Service) error {
	buildSource, _ := json.Marshal(service.BuildSource)
	resources, _ := json.Marshal(service.Resources)
	scaling, _ := json.Marshal(service.Scaling)
	healthCheck, _ := json.Marshal(service.HealthCheck)
	envVars, _ := json.Marshal(service.EnvVars)
	secretRefs, _ := json.Marshal(service.SecretRefs)
	ports, _ := json.Marshal(service.Ports)
	labels, _ := json.Marshal(service.Labels)
	annotations, _ := json.Marshal(service.Annotations)
	metadata, _ := json.Marshal(service.Metadata)
	service.UpdatedAt = time.Now()

	query := `
		UPDATE services
		SET name = $2, slug = $3, type = $4, status = $5, build_source = $6, resources = $7,
			scaling = $8, health_check = $9, env_vars = $10, secret_refs = $11, ports = $12,
			labels = $13, annotations = $14, metadata = $15, current_build_id = $16,
			current_version = $17, target_cluster_id = $18, updated_at = $19
		WHERE id = $1
	`

	result, err := r.db.pool.Exec(ctx, query,
		service.ID,
		service.Name,
		service.Slug,
		service.Type,
		service.Status,
		buildSource,
		resources,
		scaling,
		healthCheck,
		envVars,
		secretRefs,
		ports,
		labels,
		annotations,
		metadata,
		service.CurrentBuildID,
		service.CurrentVersion,
		service.TargetClusterID,
		service.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update service")
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("service", service.ID.String())
	}

	return nil
}

// Delete deletes a service
func (r *ServiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM services WHERE id = $1`

	result, err := r.db.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete service")
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("service", id.String())
	}

	return nil
}

// UpdateStatus updates only the status of a service
func (r *ServiceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ServiceStatus) error {
	query := `UPDATE services SET status = $2, updated_at = $3 WHERE id = $1`

	result, err := r.db.pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return errors.Wrap(err, "failed to update service status")
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("service", id.String())
	}

	return nil
}
