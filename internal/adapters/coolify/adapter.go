// Package coolify provides integration with Coolify for CI/Build operations.
// Coolify is used as the build system for the platform, handling Docker builds,
// buildpack builds, and managing the build pipeline.
package coolify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/northstack/platform/internal/config"
	"github.com/northstack/platform/internal/domain"
	"github.com/northstack/platform/pkg/errors"
	"github.com/northstack/platform/pkg/logger"
)

// Adapter implements the CIAdapter interface for Coolify
type Adapter struct {
	config     *config.CoolifyConfig
	httpClient *http.Client
	logger     *logger.Logger
}

// NewAdapter creates a new Coolify adapter
func NewAdapter(cfg *config.CoolifyConfig, log *logger.Logger) *Adapter {
	return &Adapter{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: log,
	}
}

// coolifyProject represents a project in Coolify's API
type coolifyProject struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	UUID        string `json:"uuid,omitempty"`
}

// coolifyApplication represents an application in Coolify's API
type coolifyApplication struct {
	ID                 string            `json:"id,omitempty"`
	Name               string            `json:"name"`
	ProjectID          string            `json:"project_id"`
	EnvironmentID      string            `json:"environment_id"`
	GitRepository      string            `json:"git_repository,omitempty"`
	GitBranch          string            `json:"git_branch,omitempty"`
	BuildPack          string            `json:"build_pack,omitempty"`
	DockerfilePath     string            `json:"dockerfile_path,omitempty"`
	DockerRegistryURL  string            `json:"docker_registry_url,omitempty"`
	DockerImageTag     string            `json:"docker_image_tag,omitempty"`
	PortsExposes       string            `json:"ports_exposes,omitempty"`
	Domains            string            `json:"domains,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
}

// coolifyBuild represents a build in Coolify's API
type coolifyBuild struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Logs        string `json:"logs,omitempty"`
	Duration    int64  `json:"duration,omitempty"`
	CreatedAt   string `json:"created_at"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	ImageTag    string `json:"image_tag,omitempty"`
	ImageDigest string `json:"image_digest,omitempty"`
	Error       string `json:"error,omitempty"`
}

// CreateProject creates a project in Coolify
func (a *Adapter) CreateProject(ctx context.Context, project *domain.Project) (string, error) {
	coolifyProj := coolifyProject{
		Name:        project.Name,
		Description: project.Description,
	}

	body, err := json.Marshal(coolifyProj)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal project")
	}

	resp, err := a.doRequest(ctx, "POST", "/api/v1/projects", body)
	if err != nil {
		return "", errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", a.handleError(resp)
	}

	var result coolifyProject
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, "failed to decode response")
	}

	a.logger.Info().
		Str("project_id", result.ID).
		Str("project_name", project.Name).
		Msg("Created project in Coolify")

	return result.ID, nil
}

// DeleteProject deletes a project from Coolify
func (a *Adapter) DeleteProject(ctx context.Context, externalID string) error {
	resp, err := a.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/projects/%s", externalID), nil)
	if err != nil {
		return errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("external_id", externalID).
		Msg("Deleted project from Coolify")

	return nil
}

// TriggerBuild triggers a new build for a service
func (a *Adapter) TriggerBuild(ctx context.Context, service *domain.Service, source domain.BuildSource) (*domain.Build, error) {
	// Prepare build request
	buildReq := map[string]interface{}{
		"application_id": service.Metadata["coolify_app_id"],
		"branch":         source.Branch,
		"commit":         source.CommitSHA,
	}

	if source.Dockerfile != "" {
		buildReq["dockerfile"] = source.Dockerfile
	}

	body, err := json.Marshal(buildReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal build request")
	}

	resp, err := a.doRequest(ctx, "POST", "/api/v1/deployments", body)
	if err != nil {
		return nil, errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, a.handleError(resp)
	}

	var result coolifyBuild
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	now := time.Now()
	build := &domain.Build{
		ID:          uuid.New(),
		ServiceID:   service.ID,
		ProjectID:   service.ProjectID,
		Status:      mapCoolifyBuildStatus(result.Status),
		Source:      source,
		TriggeredBy: "platform-orchestrator",
		CreatedAt:   now,
		Metadata: map[string]interface{}{
			"coolify_build_id": result.ID,
		},
	}

	if result.Status == "running" || result.Status == "in_progress" {
		build.StartedAt = &now
	}

	a.logger.Info().
		Str("build_id", build.ID.String()).
		Str("service_id", service.ID.String()).
		Str("coolify_build_id", result.ID).
		Msg("Build triggered in Coolify")

	return build, nil
}

// GetBuildStatus gets the current status of a build
func (a *Adapter) GetBuildStatus(ctx context.Context, buildID string) (*domain.Build, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/deployments/%s", buildID), nil)
	if err != nil {
		return nil, errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NotFound("build", buildID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, a.handleError(resp)
	}

	var result coolifyBuild
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	build := &domain.Build{
		Status:       mapCoolifyBuildStatus(result.Status),
		ImageTag:     result.ImageTag,
		ImageDigest:  result.ImageDigest,
		BuildLogs:    result.Logs,
		Duration:     result.Duration,
		ErrorMessage: result.Error,
		Metadata: map[string]interface{}{
			"coolify_build_id": result.ID,
		},
	}

	if result.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, result.StartedAt); err == nil {
			build.StartedAt = &t
		}
	}

	if result.CompletedAt != "" {
		if t, err := time.Parse(time.RFC3339, result.CompletedAt); err == nil {
			build.CompletedAt = &t
		}
	}

	return build, nil
}

// CancelBuild cancels a running build
func (a *Adapter) CancelBuild(ctx context.Context, buildID string) error {
	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/deployments/%s/cancel", buildID), nil)
	if err != nil {
		return errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("build_id", buildID).
		Msg("Build canceled in Coolify")

	return nil
}

// GetBuildLogs retrieves logs for a build
func (a *Adapter) GetBuildLogs(ctx context.Context, buildID string) (string, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/deployments/%s/logs", buildID), nil)
	if err != nil {
		return "", errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", errors.NotFound("build", buildID)
	}

	if resp.StatusCode != http.StatusOK {
		return "", a.handleError(resp)
	}

	logs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read logs")
	}

	return string(logs), nil
}

// CreateApplication creates an application in Coolify for a service
func (a *Adapter) CreateApplication(ctx context.Context, service *domain.Service, projectExternalID string, envID string) (string, error) {
	app := coolifyApplication{
		Name:          service.Name,
		ProjectID:     projectExternalID,
		EnvironmentID: envID,
	}

	// Set build configuration based on source type
	switch service.BuildSource.Type {
	case "git":
		app.GitRepository = service.BuildSource.Repository
		app.GitBranch = service.BuildSource.Branch
		if service.BuildSource.Dockerfile != "" {
			app.DockerfilePath = service.BuildSource.Dockerfile
			app.BuildPack = "dockerfile"
		} else {
			app.BuildPack = "nixpacks"
		}
	case "docker":
		app.DockerRegistryURL = service.BuildSource.Registry
		app.DockerImageTag = service.BuildSource.Image
		app.BuildPack = "dockerimage"
	}

	// Set ports
	if len(service.Ports) > 0 {
		ports := ""
		for i, port := range service.Ports {
			if i > 0 {
				ports += ","
			}
			ports += fmt.Sprintf("%d", port.Port)
		}
		app.PortsExposes = ports
	}

	// Set environment variables
	app.EnvironmentVariables = service.EnvVars

	body, err := json.Marshal(app)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal application")
	}

	resp, err := a.doRequest(ctx, "POST", "/api/v1/applications", body)
	if err != nil {
		return "", errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", a.handleError(resp)
	}

	var result coolifyApplication
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, "failed to decode response")
	}

	a.logger.Info().
		Str("app_id", result.ID).
		Str("service_id", service.ID.String()).
		Msg("Created application in Coolify")

	return result.ID, nil
}

// UpdateApplication updates an application in Coolify
func (a *Adapter) UpdateApplication(ctx context.Context, service *domain.Service, appID string) error {
	app := coolifyApplication{
		Name: service.Name,
	}

	// Update build configuration
	switch service.BuildSource.Type {
	case "git":
		app.GitRepository = service.BuildSource.Repository
		app.GitBranch = service.BuildSource.Branch
		if service.BuildSource.Dockerfile != "" {
			app.DockerfilePath = service.BuildSource.Dockerfile
		}
	case "docker":
		app.DockerRegistryURL = service.BuildSource.Registry
		app.DockerImageTag = service.BuildSource.Image
	}

	app.EnvironmentVariables = service.EnvVars

	body, err := json.Marshal(app)
	if err != nil {
		return errors.Wrap(err, "failed to marshal application")
	}

	resp, err := a.doRequest(ctx, "PATCH", fmt.Sprintf("/api/v1/applications/%s", appID), body)
	if err != nil {
		return errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("app_id", appID).
		Str("service_id", service.ID.String()).
		Msg("Updated application in Coolify")

	return nil
}

// DeleteApplication deletes an application from Coolify
func (a *Adapter) DeleteApplication(ctx context.Context, appID string) error {
	resp, err := a.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/applications/%s", appID), nil)
	if err != nil {
		return errors.DependencyFailed("coolify", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("app_id", appID).
		Msg("Deleted application from Coolify")

	return nil
}

// doRequest performs an HTTP request to the Coolify API
func (a *Adapter) doRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	url := a.config.URL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return a.httpClient.Do(req)
}

// handleError extracts error information from a response
func (a *Adapter) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	json.Unmarshal(body, &errResp)

	msg := errResp.Message
	if msg == "" {
		msg = errResp.Error
	}
	if msg == "" {
		msg = string(body)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return errors.NotFound("coolify resource", msg)
	case http.StatusUnauthorized:
		return errors.Unauthorized("invalid Coolify API key")
	case http.StatusForbidden:
		return errors.Forbidden("access denied to Coolify resource")
	case http.StatusBadRequest:
		return errors.BadRequest(msg)
	default:
		return errors.Internal(fmt.Sprintf("Coolify API error (%d): %s", resp.StatusCode, msg))
	}
}

// mapCoolifyBuildStatus maps Coolify build status to domain status
func mapCoolifyBuildStatus(status string) domain.BuildStatus {
	switch status {
	case "queued", "pending":
		return domain.BuildStatusQueued
	case "running", "in_progress", "building":
		return domain.BuildStatusRunning
	case "finished", "success", "completed":
		return domain.BuildStatusSucceeded
	case "failed", "error":
		return domain.BuildStatusFailed
	case "canceled", "cancelled":
		return domain.BuildStatusCanceled
	default:
		return domain.BuildStatusQueued
	}
}
