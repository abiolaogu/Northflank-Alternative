// Package rancher provides integration with Rancher for Kubernetes cluster management.
// Rancher is used to provision, manage, and monitor Kubernetes clusters across
// multiple cloud providers and on-premises infrastructure.
package rancher

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/northstack/platform/internal/config"
	"github.com/northstack/platform/internal/domain"
	"github.com/northstack/platform/pkg/errors"
	"github.com/northstack/platform/pkg/logger"
)

// Adapter implements the ClusterManagerAdapter interface for Rancher
type Adapter struct {
	config     *config.RancherConfig
	httpClient *http.Client
	logger     *logger.Logger
}

// NewAdapter creates a new Rancher adapter
func NewAdapter(cfg *config.RancherConfig, log *logger.Logger) *Adapter {
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
		logger: log,
	}
}

// rancherCluster represents a cluster in Rancher's API
type rancherCluster struct {
	ID                       string                 `json:"id,omitempty"`
	Name                     string                 `json:"name"`
	Description              string                 `json:"description,omitempty"`
	State                    string                 `json:"state,omitempty"`
	Provider                 string                 `json:"provider,omitempty"`
	KubernetesVersion        string                 `json:"kubernetesVersion,omitempty"`
	NodeCount                int32                  `json:"nodeCount,omitempty"`
	Transitioning            string                 `json:"transitioning,omitempty"`
	TransitioningMessage     string                 `json:"transitioningMessage,omitempty"`
	APIEndpoint              string                 `json:"apiEndpoint,omitempty"`
	Labels                   map[string]string      `json:"labels,omitempty"`
	Annotations              map[string]string      `json:"annotations,omitempty"`
	RancherKubernetesEngineConfig *rkeConfig        `json:"rancherKubernetesEngineConfig,omitempty"`
	AmazonElasticContainerServiceConfig *eksConfig  `json:"amazonElasticContainerServiceConfig,omitempty"`
	GoogleKubernetesEngineConfig *gkeConfig         `json:"googleKubernetesEngineConfig,omitempty"`
	AzureKubernetesServiceConfig *aksConfig         `json:"azureKubernetesServiceConfig,omitempty"`
}

// rkeConfig represents RKE2/K3s cluster configuration
type rkeConfig struct {
	KubernetesVersion string                   `json:"kubernetesVersion,omitempty"`
	Network           *networkConfig           `json:"network,omitempty"`
	Nodes             []rkeNode                `json:"nodes,omitempty"`
	Services          map[string]interface{}   `json:"services,omitempty"`
}

// networkConfig represents network configuration for RKE
type networkConfig struct {
	Plugin  string                 `json:"plugin,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// rkeNode represents a node in an RKE cluster
type rkeNode struct {
	Address          string   `json:"address"`
	User             string   `json:"user,omitempty"`
	Role             []string `json:"role,omitempty"`
	SSHKeyPath       string   `json:"sshKeyPath,omitempty"`
	InternalAddress  string   `json:"internalAddress,omitempty"`
}

// eksConfig represents EKS cluster configuration
type eksConfig struct {
	Region                  string   `json:"region"`
	KubernetesVersion       string   `json:"kubernetesVersion,omitempty"`
	NodeGroups              []nodeGroup `json:"nodeGroups,omitempty"`
	VPC                     string   `json:"vpc,omitempty"`
	Subnets                 []string `json:"subnets,omitempty"`
}

// gkeConfig represents GKE cluster configuration
type gkeConfig struct {
	Zone                    string   `json:"zone"`
	Region                  string   `json:"region"`
	KubernetesVersion       string   `json:"kubernetesVersion,omitempty"`
	NodePools               []nodePool `json:"nodePools,omitempty"`
	Network                 string   `json:"network,omitempty"`
	Subnetwork              string   `json:"subnetwork,omitempty"`
}

// aksConfig represents AKS cluster configuration
type aksConfig struct {
	Location                string   `json:"location"`
	KubernetesVersion       string   `json:"kubernetesVersion,omitempty"`
	NodePools               []nodePool `json:"nodePools,omitempty"`
	ResourceGroup           string   `json:"resourceGroup,omitempty"`
	VirtualNetwork          string   `json:"virtualNetwork,omitempty"`
}

// nodeGroup represents a node group for EKS
type nodeGroup struct {
	Name         string `json:"name"`
	InstanceType string `json:"instanceType,omitempty"`
	DesiredSize  int32  `json:"desiredSize,omitempty"`
	MinSize      int32  `json:"minSize,omitempty"`
	MaxSize      int32  `json:"maxSize,omitempty"`
}

// nodePool represents a node pool for GKE/AKS
type nodePool struct {
	Name              string `json:"name"`
	MachineType       string `json:"machineType,omitempty"`
	InitialNodeCount  int32  `json:"initialNodeCount,omitempty"`
	MinNodeCount      int32  `json:"minNodeCount,omitempty"`
	MaxNodeCount      int32  `json:"maxNodeCount,omitempty"`
	AutoScaling       bool   `json:"autoScaling,omitempty"`
}

// clusterCondition represents a cluster condition
type clusterCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// CreateCluster provisions a new Kubernetes cluster
func (a *Adapter) CreateCluster(ctx context.Context, cluster *domain.Cluster) (string, error) {
	rCluster := a.domainToRancher(cluster)

	body, err := json.Marshal(rCluster)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal cluster")
	}

	resp, err := a.doRequest(ctx, "POST", "/v3/clusters", body)
	if err != nil {
		return "", errors.DependencyFailed("rancher", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", a.handleError(resp)
	}

	var result rancherCluster
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, "failed to decode response")
	}

	a.logger.Info().
		Str("cluster_id", result.ID).
		Str("cluster_name", cluster.Name).
		Str("provider", string(cluster.Provider)).
		Msg("Created cluster in Rancher")

	return result.ID, nil
}

// GetCluster retrieves cluster information
func (a *Adapter) GetCluster(ctx context.Context, externalID string) (*domain.Cluster, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/v3/clusters/%s", externalID), nil)
	if err != nil {
		return nil, errors.DependencyFailed("rancher", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NotFound("cluster", externalID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, a.handleError(resp)
	}

	var rCluster rancherCluster
	if err := json.NewDecoder(resp.Body).Decode(&rCluster); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return a.rancherToDomain(&rCluster), nil
}

// UpdateCluster updates cluster configuration
func (a *Adapter) UpdateCluster(ctx context.Context, cluster *domain.Cluster) error {
	rCluster := a.domainToRancher(cluster)

	body, err := json.Marshal(rCluster)
	if err != nil {
		return errors.Wrap(err, "failed to marshal cluster")
	}

	resp, err := a.doRequest(ctx, "PUT", fmt.Sprintf("/v3/clusters/%s", cluster.RancherClusterID), body)
	if err != nil {
		return errors.DependencyFailed("rancher", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("cluster_id", cluster.RancherClusterID).
		Str("cluster_name", cluster.Name).
		Msg("Updated cluster in Rancher")

	return nil
}

// DeleteCluster deprovisions a cluster
func (a *Adapter) DeleteCluster(ctx context.Context, externalID string) error {
	resp, err := a.doRequest(ctx, "DELETE", fmt.Sprintf("/v3/clusters/%s", externalID), nil)
	if err != nil {
		return errors.DependencyFailed("rancher", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return a.handleError(resp)
	}

	a.logger.Info().
		Str("external_id", externalID).
		Msg("Deleted cluster from Rancher")

	return nil
}

// GetKubeConfig retrieves the kubeconfig for a cluster
func (a *Adapter) GetKubeConfig(ctx context.Context, externalID string) ([]byte, error) {
	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/v3/clusters/%s?action=generateKubeconfig", externalID), nil)
	if err != nil {
		return nil, errors.DependencyFailed("rancher", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NotFound("cluster", externalID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, a.handleError(resp)
	}

	var result struct {
		Config string `json:"config"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return []byte(result.Config), nil
}

// ListClusters lists all managed clusters
func (a *Adapter) ListClusters(ctx context.Context) ([]*domain.Cluster, error) {
	resp, err := a.doRequest(ctx, "GET", "/v3/clusters", nil)
	if err != nil {
		return nil, errors.DependencyFailed("rancher", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, a.handleError(resp)
	}

	var result struct {
		Data []rancherCluster `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	clusters := make([]*domain.Cluster, len(result.Data))
	for i, rc := range result.Data {
		clusters[i] = a.rancherToDomain(&rc)
	}

	return clusters, nil
}

// GetClusterHealth retrieves health status of a cluster
func (a *Adapter) GetClusterHealth(ctx context.Context, externalID string) (*domain.ClusterHealth, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/v3/clusters/%s", externalID), nil)
	if err != nil {
		return nil, errors.DependencyFailed("rancher", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NotFound("cluster", externalID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, a.handleError(resp)
	}

	var rCluster struct {
		State            string             `json:"state"`
		NodeCount        int32              `json:"nodeCount"`
		ComponentStatus  []clusterCondition `json:"componentStatuses"`
		Conditions       []clusterCondition `json:"conditions"`
		Allocatable      map[string]string  `json:"allocatable"`
		Requested        map[string]string  `json:"requested"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rCluster); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	health := &domain.ClusterHealth{
		Status:     mapRancherClusterStatus(rCluster.State),
		NodeCount:  rCluster.NodeCount,
		ReadyNodes: rCluster.NodeCount, // Simplified; would need more API calls for accurate count
		Conditions: make([]domain.ClusterCondition, len(rCluster.Conditions)),
	}

	for i, c := range rCluster.Conditions {
		health.Conditions[i] = domain.ClusterCondition{
			Type:    c.Type,
			Status:  c.Status,
			Message: c.Message,
		}
	}

	return health, nil
}

// domainToRancher converts a domain cluster to Rancher format
func (a *Adapter) domainToRancher(cluster *domain.Cluster) *rancherCluster {
	rc := &rancherCluster{
		Name:              cluster.Name,
		Description:       fmt.Sprintf("Managed by OpenPaaS - %s", cluster.Slug),
		KubernetesVersion: cluster.KubeVersion,
		Labels:            cluster.Labels,
	}

	// Configure based on provider
	switch cluster.Provider {
	case domain.ClusterProviderAWS:
		rc.AmazonElasticContainerServiceConfig = &eksConfig{
			Region:            cluster.Region,
			KubernetesVersion: cluster.KubeVersion,
			NodeGroups: []nodeGroup{
				{
					Name:        "default",
					DesiredSize: cluster.NodeCount,
					MinSize:     1,
					MaxSize:     cluster.NodeCount * 2,
				},
			},
		}
	case domain.ClusterProviderGCP:
		rc.GoogleKubernetesEngineConfig = &gkeConfig{
			Region:            cluster.Region,
			KubernetesVersion: cluster.KubeVersion,
			NodePools: []nodePool{
				{
					Name:             "default",
					InitialNodeCount: cluster.NodeCount,
					MinNodeCount:     1,
					MaxNodeCount:     cluster.NodeCount * 2,
					AutoScaling:      true,
				},
			},
		}
	case domain.ClusterProviderAzure:
		rc.AzureKubernetesServiceConfig = &aksConfig{
			Location:          cluster.Region,
			KubernetesVersion: cluster.KubeVersion,
			NodePools: []nodePool{
				{
					Name:             "default",
					InitialNodeCount: cluster.NodeCount,
					MinNodeCount:     1,
					MaxNodeCount:     cluster.NodeCount * 2,
					AutoScaling:      true,
				},
			},
		}
	case domain.ClusterProviderK3s, domain.ClusterProviderOnPrem:
		rc.RancherKubernetesEngineConfig = &rkeConfig{
			KubernetesVersion: cluster.KubeVersion,
			Network: &networkConfig{
				Plugin: "canal",
			},
		}
	}

	return rc
}

// rancherToDomain converts a Rancher cluster to domain format
func (a *Adapter) rancherToDomain(rc *rancherCluster) *domain.Cluster {
	cluster := &domain.Cluster{
		Name:             rc.Name,
		RancherClusterID: rc.ID,
		Status:           mapRancherClusterStatus(rc.State),
		KubeVersion:      rc.KubernetesVersion,
		NodeCount:        rc.NodeCount,
		APIEndpoint:      rc.APIEndpoint,
		Labels:           rc.Labels,
	}

	// Determine provider and region from config
	if rc.AmazonElasticContainerServiceConfig != nil {
		cluster.Provider = domain.ClusterProviderAWS
		cluster.Region = rc.AmazonElasticContainerServiceConfig.Region
	} else if rc.GoogleKubernetesEngineConfig != nil {
		cluster.Provider = domain.ClusterProviderGCP
		cluster.Region = rc.GoogleKubernetesEngineConfig.Region
	} else if rc.AzureKubernetesServiceConfig != nil {
		cluster.Provider = domain.ClusterProviderAzure
		cluster.Region = rc.AzureKubernetesServiceConfig.Location
	} else if rc.RancherKubernetesEngineConfig != nil {
		cluster.Provider = domain.ClusterProviderK3s
	} else {
		cluster.Provider = domain.ClusterProviderOnPrem
	}

	return cluster
}

// doRequest performs an HTTP request to the Rancher API
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

	// Set authentication
	if a.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.Token)
	} else if a.config.AccessKey != "" && a.config.SecretKey != "" {
		req.SetBasicAuth(a.config.AccessKey, a.config.SecretKey)
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
		Code    string `json:"code"`
		Status  int    `json:"status"`
	}
	json.Unmarshal(body, &errResp)

	msg := errResp.Message
	if msg == "" {
		msg = string(body)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return errors.NotFound("rancher resource", msg)
	case http.StatusUnauthorized:
		return errors.Unauthorized("invalid Rancher credentials")
	case http.StatusForbidden:
		return errors.Forbidden("access denied to Rancher resource")
	case http.StatusBadRequest:
		return errors.BadRequest(msg)
	default:
		return errors.Internal(fmt.Sprintf("Rancher API error (%d): %s", resp.StatusCode, msg))
	}
}

// mapRancherClusterStatus maps Rancher cluster state to domain status
func mapRancherClusterStatus(state string) domain.ClusterStatus {
	switch state {
	case "provisioning", "pending", "waiting":
		return domain.ClusterStatusProvisioning
	case "active", "running":
		return domain.ClusterStatusActive
	case "upgrading", "updating":
		return domain.ClusterStatusUpgrading
	case "error", "unavailable", "degraded":
		return domain.ClusterStatusUnhealthy
	case "removing", "removed":
		return domain.ClusterStatusDeleting
	default:
		return domain.ClusterStatusProvisioning
	}
}
