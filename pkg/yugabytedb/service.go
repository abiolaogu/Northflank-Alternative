package yugabytedb

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// YugabyteDB Kubernetes operator GVRs
var (
	YBClusterGVR = schema.GroupVersionResource{
		Group:    "yugabyte.com",
		Version:  "v1alpha1",
		Resource: "ybclusters",
	}
)

// DatabaseService manages YugabyteDB clusters via Kubernetes operator
type DatabaseService struct {
	dynamic   dynamic.Interface
	namespace string
}

// CreateDatabaseInput holds parameters for creating a database cluster
type CreateDatabaseInput struct {
	Name             string
	ProjectID        string
	TeamID           string
	Size             string // small, medium, large, xlarge
	StorageGB        int
	TServerReplicas  int
	MasterReplicas   int
	HighAvailability bool
	BackupEnabled    bool
	TLSEnabled       bool
	Version          string
}

// DatabaseInfo holds database connection information
type DatabaseInfo struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	YSQLEndpoint    string     `json:"ysql_endpoint"` // PostgreSQL-compatible
	YCQLEndpoint    string     `json:"ycql_endpoint"` // Cassandra-compatible
	MasterUIURL     string     `json:"master_ui_url"`
	Port            int        `json:"port"`
	Database        string     `json:"database"`
	Username        string     `json:"username"`
	SecretName      string     `json:"secret_name"`
	TServerReplicas int        `json:"tserver_replicas"`
	MasterReplicas  int        `json:"master_replicas"`
	ReadyTServers   int        `json:"ready_tservers"`
	ReadyMasters    int        `json:"ready_masters"`
	StorageUsedGB   float64    `json:"storage_used_gb"`
	LastBackupTime  *time.Time `json:"last_backup_time,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ResourceConfig holds resource configuration for different sizes
type ResourceConfig struct {
	TServerCPURequest    string
	TServerCPULimit      string
	TServerMemoryRequest string
	TServerMemoryLimit   string
	MasterCPURequest     string
	MasterCPULimit       string
	MasterMemoryRequest  string
	MasterMemoryLimit    string
}

// NewDatabaseService creates a new YugabyteDB service
func NewDatabaseService(dynClient dynamic.Interface, namespace string) *DatabaseService {
	return &DatabaseService{
		dynamic:   dynClient,
		namespace: namespace,
	}
}

// CreateDatabase creates a new YugabyteDB cluster
func (s *DatabaseService) CreateDatabase(ctx context.Context, input *CreateDatabaseInput) (*DatabaseInfo, error) {
	clusterName := fmt.Sprintf("%s-%s", input.ProjectID, input.Name)
	secretName := fmt.Sprintf("%s-credentials", clusterName)

	// Map size to resources
	resources := s.mapSizeToResources(input.Size)

	// Determine replica counts
	tserverReplicas := 3
	masterReplicas := 3
	if !input.HighAvailability {
		tserverReplicas = 1
		masterReplicas = 1
	}
	if input.TServerReplicas > 0 {
		tserverReplicas = input.TServerReplicas
	}
	if input.MasterReplicas > 0 {
		masterReplicas = input.MasterReplicas
	}

	// YugabyteDB version
	version := input.Version
	if version == "" {
		version = "2.20.1.0-b97"
	}

	// Build cluster spec
	cluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "yugabyte.com/v1alpha1",
			"kind":       "YBCluster",
			"metadata": map[string]interface{}{
				"name":      clusterName,
				"namespace": s.namespace,
				"labels": map[string]interface{}{
					"northstack.io/project": input.ProjectID,
					"northstack.io/team":    input.TeamID,
					"northstack.io/type":    "database",
				},
			},
			"spec": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": "yugabytedb/yugabyte",
					"tag":        version,
					"pullPolicy": "IfNotPresent",
				},
				"tserver": map[string]interface{}{
					"replicas": tserverReplicas,
					"storage": map[string]interface{}{
						"count": 1,
						"size":  fmt.Sprintf("%dGi", input.StorageGB),
					},
					"resource": map[string]interface{}{
						"requests": map[string]interface{}{
							"cpu":    resources.TServerCPURequest,
							"memory": resources.TServerMemoryRequest,
						},
						"limits": map[string]interface{}{
							"cpu":    resources.TServerCPULimit,
							"memory": resources.TServerMemoryLimit,
						},
					},
					"gflags": map[string]interface{}{
						"enable_ysql":          "true",
						"ysql_enable_auth":     "true",
						"ysql_max_connections": "300",
					},
				},
				"master": map[string]interface{}{
					"replicas": masterReplicas,
					"storage": map[string]interface{}{
						"count": 1,
						"size":  "10Gi",
					},
					"resource": map[string]interface{}{
						"requests": map[string]interface{}{
							"cpu":    resources.MasterCPURequest,
							"memory": resources.MasterMemoryRequest,
						},
						"limits": map[string]interface{}{
							"cpu":    resources.MasterCPULimit,
							"memory": resources.MasterMemoryLimit,
						},
					},
					"gflags": map[string]interface{}{
						"replication_factor":    fmt.Sprintf("%d", tserverReplicas),
						"enable_load_balancing": "true",
					},
				},
				"tls": map[string]interface{}{
					"enabled": input.TLSEnabled,
				},
				"enableBackup": input.BackupEnabled,
			},
		},
	}

	// Create credentials secret first
	if err := s.createCredentialsSecret(ctx, secretName, input.Name); err != nil {
		return nil, fmt.Errorf("failed to create credentials secret: %w", err)
	}

	// Create the cluster
	_, err := s.dynamic.Resource(YBClusterGVR).Namespace(s.namespace).Create(ctx, cluster, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create YugabyteDB cluster: %w", err)
	}

	return &DatabaseInfo{
		ID:              clusterName,
		Name:            input.Name,
		Status:          "creating",
		YSQLEndpoint:    fmt.Sprintf("%s-yb-tserver-service.%s.svc:5433", clusterName, s.namespace),
		YCQLEndpoint:    fmt.Sprintf("%s-yb-tserver-service.%s.svc:9042", clusterName, s.namespace),
		Port:            5433,
		Database:        input.Name,
		Username:        input.Name,
		SecretName:      secretName,
		TServerReplicas: tserverReplicas,
		MasterReplicas:  masterReplicas,
		CreatedAt:       time.Now(),
	}, nil
}

// GetDatabase retrieves database information
func (s *DatabaseService) GetDatabase(ctx context.Context, name string) (*DatabaseInfo, error) {
	cluster, err := s.dynamic.Resource(YBClusterGVR).Namespace(s.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return s.extractDatabaseInfo(cluster), nil
}

// ListDatabases lists all YugabyteDB clusters
func (s *DatabaseService) ListDatabases(ctx context.Context, projectID string) ([]*DatabaseInfo, error) {
	labelSelector := ""
	if projectID != "" {
		labelSelector = fmt.Sprintf("northstack.io/project=%s", projectID)
	}

	clusters, err := s.dynamic.Resource(YBClusterGVR).Namespace(s.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	var databases []*DatabaseInfo
	for _, cluster := range clusters.Items {
		databases = append(databases, s.extractDatabaseInfo(&cluster))
	}

	return databases, nil
}

// DeleteDatabase deletes a YugabyteDB cluster
func (s *DatabaseService) DeleteDatabase(ctx context.Context, name string) error {
	return s.dynamic.Resource(YBClusterGVR).Namespace(s.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ScaleDatabase scales the TServer replicas
func (s *DatabaseService) ScaleDatabase(ctx context.Context, name string, replicas int) error {
	cluster, err := s.dynamic.Resource(YBClusterGVR).Namespace(s.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	spec := cluster.Object["spec"].(map[string]interface{})
	tserver := spec["tserver"].(map[string]interface{})
	tserver["replicas"] = replicas

	_, err = s.dynamic.Resource(YBClusterGVR).Namespace(s.namespace).Update(ctx, cluster, metav1.UpdateOptions{})
	return err
}

func (s *DatabaseService) mapSizeToResources(size string) ResourceConfig {
	configs := map[string]ResourceConfig{
		"small": {
			TServerCPURequest:    "500m",
			TServerCPULimit:      "1",
			TServerMemoryRequest: "1Gi",
			TServerMemoryLimit:   "2Gi",
			MasterCPURequest:     "250m",
			MasterCPULimit:       "500m",
			MasterMemoryRequest:  "512Mi",
			MasterMemoryLimit:    "1Gi",
		},
		"medium": {
			TServerCPURequest:    "1",
			TServerCPULimit:      "2",
			TServerMemoryRequest: "2Gi",
			TServerMemoryLimit:   "4Gi",
			MasterCPURequest:     "500m",
			MasterCPULimit:       "1",
			MasterMemoryRequest:  "1Gi",
			MasterMemoryLimit:    "2Gi",
		},
		"large": {
			TServerCPURequest:    "2",
			TServerCPULimit:      "4",
			TServerMemoryRequest: "4Gi",
			TServerMemoryLimit:   "8Gi",
			MasterCPURequest:     "1",
			MasterCPULimit:       "2",
			MasterMemoryRequest:  "2Gi",
			MasterMemoryLimit:    "4Gi",
		},
		"xlarge": {
			TServerCPURequest:    "4",
			TServerCPULimit:      "8",
			TServerMemoryRequest: "8Gi",
			TServerMemoryLimit:   "16Gi",
			MasterCPURequest:     "2",
			MasterCPULimit:       "4",
			MasterMemoryRequest:  "4Gi",
			MasterMemoryLimit:    "8Gi",
		},
	}

	if config, ok := configs[size]; ok {
		return config
	}
	return configs["small"]
}

func (s *DatabaseService) createCredentialsSecret(ctx context.Context, name, username string) error {
	// Generate random password
	password := uuid.New().String()

	secret := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": s.namespace,
			},
			"type": "Opaque",
			"stringData": map[string]interface{}{
				"username": username,
				"password": password,
				"database": username,
				"ysql_uri": fmt.Sprintf("postgresql://%s:%s@yugabyte-ysql.%s.svc:5433/%s", username, password, s.namespace, username),
			},
		},
	}

	secretGVR := schema.GroupVersionResource{Version: "v1", Resource: "secrets"}
	_, err := s.dynamic.Resource(secretGVR).Namespace(s.namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func (s *DatabaseService) extractDatabaseInfo(cluster *unstructured.Unstructured) *DatabaseInfo {
	info := &DatabaseInfo{
		ID:   cluster.GetName(),
		Name: cluster.GetName(),
	}

	spec, ok := cluster.Object["spec"].(map[string]interface{})
	if ok {
		if tserver, ok := spec["tserver"].(map[string]interface{}); ok {
			if replicas, ok := tserver["replicas"].(int64); ok {
				info.TServerReplicas = int(replicas)
			}
		}
		if master, ok := spec["master"].(map[string]interface{}); ok {
			if replicas, ok := master["replicas"].(int64); ok {
				info.MasterReplicas = int(replicas)
			}
		}
	}

	status, ok := cluster.Object["status"].(map[string]interface{})
	if ok {
		if phase, ok := status["phase"].(string); ok {
			info.Status = phase
		}
	}

	info.YSQLEndpoint = fmt.Sprintf("%s-yb-tserver-service.%s.svc:5433", info.Name, s.namespace)
	info.YCQLEndpoint = fmt.Sprintf("%s-yb-tserver-service.%s.svc:9042", info.Name, s.namespace)
	info.Port = 5433

	return info
}
