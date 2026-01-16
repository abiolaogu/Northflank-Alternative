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

// ProjectRepository implements domain.ProjectRepository using PostgreSQL
type ProjectRepository struct {
	db *PostgresDB
}

// NewProjectRepository creates a new ProjectRepository
func NewProjectRepository(db *PostgresDB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	labels, _ := json.Marshal(project.Labels)
	metadata, _ := json.Marshal(project.Metadata)

	query := `
		INSERT INTO projects (id, name, slug, description, status, owner_id, team_id, labels, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.pool.Exec(ctx, query,
		project.ID,
		project.Name,
		project.Slug,
		project.Description,
		project.Status,
		project.OwnerID,
		project.TeamID,
		labels,
		metadata,
		project.CreatedAt,
		project.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, "failed to create project")
	}

	return nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT id, name, slug, description, status, owner_id, team_id, labels, metadata, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	project := &domain.Project{}
	var labels, metadata []byte

	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.Status,
		&project.OwnerID,
		&project.TeamID,
		&labels,
		&metadata,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("project", id.String())
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	json.Unmarshal(labels, &project.Labels)
	json.Unmarshal(metadata, &project.Metadata)

	return project, nil
}

// GetBySlug retrieves a project by slug
func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	query := `
		SELECT id, name, slug, description, status, owner_id, team_id, labels, metadata, created_at, updated_at
		FROM projects
		WHERE slug = $1
	`

	project := &domain.Project{}
	var labels, metadata []byte

	err := r.db.pool.QueryRow(ctx, query, slug).Scan(
		&project.ID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.Status,
		&project.OwnerID,
		&project.TeamID,
		&labels,
		&metadata,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("project", slug)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	json.Unmarshal(labels, &project.Labels)
	json.Unmarshal(metadata, &project.Metadata)

	return project, nil
}

// List retrieves projects with optional filtering
func (r *ProjectRepository) List(ctx context.Context, filter domain.ProjectFilter) ([]*domain.Project, error) {
	query := `
		SELECT id, name, slug, description, status, owner_id, team_id, labels, metadata, created_at, updated_at
		FROM projects
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.OwnerID != nil {
		query += fmt.Sprintf(" AND owner_id = $%d", argIndex)
		args = append(args, *filter.OwnerID)
		argIndex++
	}

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND team_id = $%d", argIndex)
		args = append(args, *filter.TeamID)
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
		return nil, errors.Wrap(err, "failed to list projects")
	}
	defer rows.Close()

	projects := []*domain.Project{}
	for rows.Next() {
		project := &domain.Project{}
		var labels, metadata []byte

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Slug,
			&project.Description,
			&project.Status,
			&project.OwnerID,
			&project.TeamID,
			&labels,
			&metadata,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan project")
		}

		json.Unmarshal(labels, &project.Labels)
		json.Unmarshal(metadata, &project.Metadata)

		projects = append(projects, project)
	}

	return projects, nil
}

// Update updates an existing project
func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	labels, _ := json.Marshal(project.Labels)
	metadata, _ := json.Marshal(project.Metadata)
	project.UpdatedAt = time.Now()

	query := `
		UPDATE projects
		SET name = $2, slug = $3, description = $4, status = $5, team_id = $6, labels = $7, metadata = $8, updated_at = $9
		WHERE id = $1
	`

	result, err := r.db.pool.Exec(ctx, query,
		project.ID,
		project.Name,
		project.Slug,
		project.Description,
		project.Status,
		project.TeamID,
		labels,
		metadata,
		project.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("project", project.ID.String())
	}

	return nil
}

// Delete deletes a project
func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := r.db.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete project")
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("project", id.String())
	}

	return nil
}
