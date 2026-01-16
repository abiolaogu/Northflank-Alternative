// Package argocd provides integration with ArgoCD for GitOps-based deployments.
// ArgoCD handles the continuous delivery aspect of the platform, managing
// Kubernetes manifests and ensuring cluster state matches the desired state.
package argocd

import (
	"bytes"
	"context"
	"crypto/tls"
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

// Adapter implements the GitOpsAdapter interface for ArgoCD
type Adapter struct {
	config     *config.ArgoCDConfig
	httpClient *http.Client
	logger     *logger.Logger
	authToken  string
}

// NewAdapter creates a new ArgoCD adapter
func NewAdapter(cfg *config.ArgoCDConfig, log *logger.Logger) *Adapter {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.TLSSkipVerify,
		},
	}

	return &Adapter{
		config: cfg,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		logger:    log,
		authToken: cfg.Token,
	}
}

// argoApplication represents an ArgoCD application
type argoApplication struct {
	Metadata argoMetadata      `json:"metadata"`
	Spec     argoApplicationSpec `json:"spec"`
	Status   *argoApplicationStatus `json:"status,omitempty"`
}

// argoMetadata represents ArgoCD resource metadata
type argoMetadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// argoApplicationSpec represents ArgoCD application spec
type argoApplicationSpec struct {
	Project     string          `json:"project"`
	Source      argoSource      `json:"source"`
	Destination argoDestination `json:"destination"`
	SyncPolicy  *argoSyncPolicy `json:"syncPolicy,omitempty"`
}

// argoSource represents the application source
type argoSource struct {
	RepoURL        string      `json:"repoURL"`
	Path           string      `json:"path,omitempty"`
	TargetRevision string      `json:"targetRevision"`
	Helm           *argoHelm   `json:"helm,omitempty"`
	Kustomize      *argoKustomize `json:"kustomize,omitempty"`
}

// argoHelm represents Helm-specific configuration
type argoHelm struct {
	ReleaseName string            `json:"releaseName,omitempty"`
	ValueFiles  []string          `json:"valueFiles,omitempty"`
	Values      string            `json:"values,omitempty"`
	Parameters  []argoHelmParam   `json:"parameters,omitempty"`
}

// argoHelmParam represents a Helm parameter
type argoHelmParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// argoKustomize represents Kustomize-specific configuration
type argoKustomize struct {
	Images []string `json:"images,omitempty"`
}

// argoDestination represents the deployment destination
type argoDestination struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

// argoSyncPolicy represents the sync policy
type argoSyncPolicy struct {
	Automated      *argoAutomatedSync `json:"automated,omitempty"`
	SyncOptions    []string           `json:"syncOptions,omitempty"`
	Retry          *argoRetry         `json:"retry,omitempty"`
}

// argoAutomatedSync represents automated sync configuration
type argoAutomatedSync struct {
	Prune      bool `json:"prune"`
	SelfHeal   bool `json:"selfHeal"`
	AllowEmpty bool `json:"allowEmpty"`
}

// argoRetry represents retry configuration
type argoRetry struct {
	Limit   int64        `json:"limit"`
	Backoff argoBackoff  `json:"backoff"`
}

// argoBackoff represents backoff configuration
type argoBackoff struct {
	Duration    string `json:"duration"`
	Factor      int64  `json:"factor"`
	MaxDuration string `json:"maxDuration"`
}

// argoApplicationStatus represents application status
type argoApplicationStatus struct {
	Health           argoHealthStatus   `json:"health"`
	Sync             argoSyncStatus     `json:"sync"`
	OperationState   *argoOperationState `json:"operationState,omitempty"`
	Resources        []argoResourceStatus `json:"resources,omitempty"`
	Summary          argoSummary        `json:"summary"`
	History          []argoRevisionHistory `json:"history,omitempty"`
}

// argoHealthStatus represents health status
type argoHealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// argoSyncStatus represents sync status
type argoSyncStatus struct {
	Status    string `json:"status"`
	Revision  string `json:"revision,omitempty"`
	ComparedTo struct {
		Source      argoSource      `json:"source"`
		Destination argoDestination `json:"destination"`
	} `json:"comparedTo"`
}

// argoOperationState represents operation state
type argoOperationState struct {
	Phase     string    `json:"phase"`
	Message   string    `json:"message,omitempty"`
	StartedAt string    `json:"startedAt,omitempty"`
	FinishedAt string   `json:"finishedAt,omitempty"`
}

// argoResourceStatus represents a resource status
type argoResourceStatus struct {
	Group     string `json:"group,omitempty"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Health    *argoHealthStatus `json:"health,omitempty"`
}

// argoSummary represents application summary
type argoSummary struct {
	Images []string `json:"images,omitempty"`
}

// argoRevisionHistory represents revision history
type argoRevisionHistory struct {
	Revision     string    `json:"revision"`
	DeployedAt   string    `json:"deployedAt"`
	ID           int64     `json:"id"`
	Source       argoSource `json:"source"`
}

// Authenticate authenticates with ArgoCD and gets a session token
func (a *Adapter) Authenticate(ctx context.Context) error {
	if a.config.Token != "" {
		a.authToken = a.config.Token
		return nil
	}

	authReq := map[string]string{
		"username": a.config.Username,
		"password": a.config.Password,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return errors.Wrap(err, "failed to marshal auth request")
	}

	resp, err := a.doRequest(ctx, "POST", "/api/v1/session", body, false)
	if err != nil {
		return errors.DependencyFailed("argocd", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return errors.Wrap(err, "failed to decode response")
	}

	a.authToken = result.Token
	a.logger.Info().Msg("Authenticated with ArgoCD")

	return nil
}

// CreateApplication creates a new application in ArgoCD
func (a *Adapter) CreateApplication(ctx context.Context, service *domain.Service, environment *domain.Environment) (string, error) {
	appName := fmt.Sprintf("%s-%s", service.Slug, environment.Slug)

	app := argoApplication{
		Metadata: argoMetadata{
			Name:      appName,
			Namespace: "argocd",
			Labels: map[string]string{
				"openpaas.io/service-id":     service.ID.String(),
				"openpaas.io/project-id":     service.ProjectID.String(),
				"openpaas.io/environment-id": environment.ID.String(),
			},
			Annotations: service.Annotations,
		},
		Spec: argoApplicationSpec{
			Project: a.config.AppProject,
			Source: argoSource{
				RepoURL:        a.config.RepoURL,
				Path:           fmt.Sprintf("services/%s/%s", service.Slug, environment.Slug),
				TargetRevision: a.config.TargetRevision,
			},
			Destination: argoDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: environment.Namespace,
			},
		},
	}

	// Set sync policy based on config
	if a.config.SyncPolicy.Automated {
		app.Spec.SyncPolicy = &argoSyncPolicy{
			Automated: &argoAutomatedSync{
				Prune:      a.config.SyncPolicy.Prune,
				SelfHeal:   a.config.SyncPolicy.SelfHeal,
				AllowEmpty: a.config.SyncPolicy.AllowEmpty,
			},
			SyncOptions: []string{"CreateNamespace=true"},
			Retry: &argoRetry{
				Limit: 5,
				Backoff: argoBackoff{
					Duration:    "5s",
					Factor:      2,
					MaxDuration: "3m",
				},
			},
		}
	}

	// Handle Kustomize for image updates
	if service.CurrentVersion != "" {
		app.Spec.Source.Kustomize = &argoKustomize{
			Images: []string{
				fmt.Sprintf("%s:%s", service.BuildSource.Image, service.CurrentVersion),
			},
		}
	}

	body, err := json.Marshal(app)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal application")
	}

	resp, err := a.doRequest(ctx, "POST", "/api/v1/applications", body, true)
	if err != nil {
		return "", errors.DependencyFailed("argocd", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", a.handleError(resp)
	}

	var result argoApplication
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, "failed to decode response")
	}

	a.logger.Info().
		Str("app_name", appName).
		Str("service_id", service.ID.String()).
		Str("environment", environment.Name).
		Msg("Created application in ArgoCD")

	return appName, nil
}

// UpdateApplication updates an existing ArgoCD application
func (a *Adapter) UpdateApplication(ctx context.Context, service *domain.Service, environment *domain.Environment) error {
	appName := fmt.Sprintf("%s-%s", service.Slug, environment.Slug)

	// Get existing application
	existing, err := a.getApplication(ctx, appName)
	if err != nil {
		return err
	}

	// Update image if version changed
	if service.CurrentVersion != "" {
		if existing.Spec.Source.Kustomize == nil {
			existing.Spec.Source.Kustomize = &argoKustomize{}
		}
		existing.Spec.Source.Kustomize.Images = []string{
			fmt.Sprintf("%s:%s", service.BuildSource.Image, service.CurrentVersion),
		}
	}

	// Update labels
	existing.Metadata.Labels["openpaas.io/version"] = service.CurrentVersion

	body, err := json.Marshal(existing)
	if err != nil {
		return errors.Wrap(err, "failed to marshal application")
	}

	resp, err := a.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/applications/%s", appName), body, true)
	if err != nil {
		return errors.DependencyFailed("argocd", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("app_name", appName).
		Str("version", service.CurrentVersion).
		Msg("Updated application in ArgoCD")

	return nil
}

// DeleteApplication removes an application from ArgoCD
func (a *Adapter) DeleteApplication(ctx context.Context, externalID string) error {
	resp, err := a.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/applications/%s?cascade=true", externalID), nil, true)
	if err != nil {
		return errors.DependencyFailed("argocd", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("app_name", externalID).
		Msg("Deleted application from ArgoCD")

	return nil
}

// SyncApplication triggers a sync for an application
func (a *Adapter) SyncApplication(ctx context.Context, externalID string) error {
	syncReq := map[string]interface{}{
		"prune":    true,
		"strategy": map[string]interface{}{"apply": map[string]bool{"force": false}},
	}

	body, err := json.Marshal(syncReq)
	if err != nil {
		return errors.Wrap(err, "failed to marshal sync request")
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/applications/%s/sync", externalID), body, true)
	if err != nil {
		return errors.DependencyFailed("argocd", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("app_name", externalID).
		Msg("Triggered sync in ArgoCD")

	return nil
}

// GetApplicationStatus retrieves the status of an application
func (a *Adapter) GetApplicationStatus(ctx context.Context, externalID string) (*domain.ApplicationStatus, error) {
	app, err := a.getApplication(ctx, externalID)
	if err != nil {
		return nil, err
	}

	status := &domain.ApplicationStatus{
		Health:     app.Status.Health.Status,
		SyncStatus: app.Status.Sync.Status,
	}

	// Extract current and desired images
	if len(app.Status.Summary.Images) > 0 {
		status.CurrentImage = app.Status.Summary.Images[0]
	}
	if app.Spec.Source.Kustomize != nil && len(app.Spec.Source.Kustomize.Images) > 0 {
		status.DesiredImage = app.Spec.Source.Kustomize.Images[0]
	}

	// Map resources
	status.Resources = make([]domain.ResourceStatus, len(app.Status.Resources))
	for i, r := range app.Status.Resources {
		rs := domain.ResourceStatus{
			Kind:      r.Kind,
			Name:      r.Name,
			Namespace: r.Namespace,
			Status:    r.Status,
		}
		if r.Health != nil {
			rs.Health = r.Health.Status
			rs.Message = r.Health.Message
		}
		status.Resources[i] = rs

		// Count replicas from Deployment resources
		if r.Kind == "Deployment" && r.Health != nil && r.Health.Status == "Healthy" {
			status.ReadyReplicas++
		}
		if r.Kind == "Deployment" {
			status.Replicas++
		}
	}

	return status, nil
}

// GetApplicationHistory retrieves deployment history
func (a *Adapter) GetApplicationHistory(ctx context.Context, externalID string) ([]*domain.Deployment, error) {
	app, err := a.getApplication(ctx, externalID)
	if err != nil {
		return nil, err
	}

	deployments := make([]*domain.Deployment, len(app.Status.History))
	for i, h := range app.Status.History {
		deployedAt, _ := time.Parse(time.RFC3339, h.DeployedAt)
		deployments[i] = &domain.Deployment{
			ID:          uuid.New(),
			Version:     h.Revision,
			Status:      domain.DeploymentStatusSucceeded,
			CompletedAt: &deployedAt,
			Metadata: map[string]interface{}{
				"argocd_revision_id": h.ID,
				"source_path":        h.Source.Path,
			},
		}
	}

	return deployments, nil
}

// RollbackApplication rolls back to a previous version
func (a *Adapter) RollbackApplication(ctx context.Context, externalID string, revision int64) error {
	rollbackReq := map[string]interface{}{
		"id":       revision,
		"prune":    true,
	}

	body, err := json.Marshal(rollbackReq)
	if err != nil {
		return errors.Wrap(err, "failed to marshal rollback request")
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/applications/%s/rollback", externalID), body, true)
	if err != nil {
		return errors.DependencyFailed("argocd", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("app_name", externalID).
		Int64("revision", revision).
		Msg("Rolled back application in ArgoCD")

	return nil
}

// getApplication retrieves an application from ArgoCD
func (a *Adapter) getApplication(ctx context.Context, name string) (*argoApplication, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/applications/%s", name), nil, true)
	if err != nil {
		return nil, errors.DependencyFailed("argocd", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NotFound("application", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, a.handleError(resp)
	}

	var app argoApplication
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &app, nil
}

// doRequest performs an HTTP request to the ArgoCD API
func (a *Adapter) doRequest(ctx context.Context, method, path string, body []byte, auth bool) (*http.Response, error) {
	url := a.config.URL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if auth && a.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.authToken)
	}
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
		return errors.NotFound("argocd resource", msg)
	case http.StatusUnauthorized:
		return errors.Unauthorized("invalid ArgoCD credentials")
	case http.StatusForbidden:
		return errors.Forbidden("access denied to ArgoCD resource")
	case http.StatusBadRequest:
		return errors.BadRequest(msg)
	default:
		return errors.Internal(fmt.Sprintf("ArgoCD API error (%d): %s", resp.StatusCode, msg))
	}
}
