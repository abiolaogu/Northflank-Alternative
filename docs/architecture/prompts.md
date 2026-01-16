#   
  
  
Here are comprehensive prompts for Claude Code to build this platform:  
**Prompt 1: Project Initialization and Core Structure**  
  
  
markdown  
# NorthStack Platform - Project Initialization  
  
## Context  
I'm building an open-source alternative to Northflank - a unified PaaS platform that orchestrates Coolify (CI/CD), Rancher (cluster management), ArgoCD (GitOps), with YugabyteDB (distributed SQL), DragonflyDB (Redis-compatible cache), NATS JetStream (events), Hasura (GraphQL), and RKE2 (Kubernetes).  
  
## Task  
Initialize the complete Go project structure with proper module organization, configuration management, and core domain models.  
  
## Requirements  
  
### 1. Project Structure  
Create this directory layout:  
```  
northstack/  
├── cmd/  
│   ├── api/              # Main API server  
│   │   └── main.go  
│   ├── worker/           # Background worker  
│   │   └── main.go  
│   └── cli/              # CLI tool  
│       └── main.go  
├── internal/  
│   ├── config/           # Configuration management  
│   ├── models/           # Domain models  
│   ├── database/         # YugabyteDB client  
│   ├── cache/            # DragonflyDB client  
│   ├── events/           # NATS event bus  
│   ├── api/              # HTTP handlers  
│   ├── services/         # Business logic  
│   ├── middleware/       # HTTP middleware  
│   └── workers/          # Background jobs  
├── pkg/  
│   ├── coolify/          # Coolify API client  
│   ├── rancher/          # Rancher API client  
│   ├── argocd/           # ArgoCD client  
│   ├── hasura/           # Hasura client  
│   ├── rke2/             # RKE2 manager  
│   └── vault/            # Vault client  
├── migrations/           # Database migrations  
├── deployments/  
│   ├── docker/           # Docker compose files  
│   ├── kubernetes/       # K8s manifests  
│   └── helm/             # Helm charts  
├── scripts/              # Utility scripts  
├── config/               # Configuration templates  
├── docs/                 # Documentation  
├── go.mod  
├── go.sum  
├── Makefile  
└── .env.example  
```  
  
### 2. Go Module Setup  
- Module name: `github.com/northstack/platform`  
- Go version: 1.22+  
- Key dependencies:  
  - github.com/gofiber/fiber/v2 (HTTP framework)  
  - github.com/jackc/pgx/v5 (PostgreSQL/YugabyteDB driver)  
  - github.com/redis/go-redis/v9 (DragonflyDB client)  
  - github.com/nats-io/nats.go (NATS client)  
  - github.com/spf13/viper (configuration)  
  - go.uber.org/zap (logging)  
  - github.com/golang-migrate/migrate/v4 (migrations)  
  
### 3. Configuration System  
Create a comprehensive config struct that supports:  
- Environment variables with NORTHSTACK_ prefix  
- YAML/JSON config files  
- Default values  
- Validation  
- Hot reloading  
  
Config sections needed:  
- Server (port, host, timeouts)  
- YugabyteDB (hosts, credentials, pool settings, load balancing)  
- DragonflyDB (addresses, password, pool settings)  
- NATS (URL, credentials, JetStream settings)  
- Hasura (endpoint, admin secret)  
- Coolify (base URL, API key)  
- Rancher (server URL, access/secret keys)  
- ArgoCD (server, auth token)  
- Auth (JWT secret, token expiry)  
- Observability (Prometheus, tracing)  
  
### 4. Domain Models  
Create models for:  
- Project (id, name, description, team_id, settings)  
- Application (id, project_id, name, repository, branch, build_pack, environment)  
- Deployment (id, application_id, revision, status, strategy, timestamps)  
- Cluster (id, name, provider, version, status, node_count)  
- Team (id, name, members)  
- User (id, email, name, roles)  
- AuditLog (id, actor, action, resource, timestamp, details)  
  
Use proper JSON tags, database tags, and validation tags.  
  
### 5. Makefile  
### 5. Makefile  
Create targets for:  
- build (all binaries)  
- test (with coverage)  
- lint (golangci-lint)  
- migrate-up/migrate-down  
- docker-build  
- docker-up/docker-down  
- generate (for any code generation)  
- clean  
  
## Deliverables  
1. Complete directory structure with placeholder files  
2. go.mod with all dependencies  
3. internal/config/config.go with full configuration system  
4. internal/models/models.go with all domain models  
5. Makefile with all targets  
6. .env.example with all environment variables  
7. config/config.example.yaml with documented settings  
  
## Code Quality Requirements  
## Code Quality Requirements  
- Use Go best practices and idioms  
- Comprehensive error handling with wrapped errors  
- Full documentation comments  
- Thread-safe where applicable  
- Testable design (interfaces for dependencies)  
  
**Prompt 2: YugabyteDB and DragonflyDB Integration**  
  
  
markdown  
# NorthStack Platform - Data Layer Implementation  
  
## Context  
## Context  
Building the data layer for NorthStack platform using YugabyteDB (distributed SQL) and DragonflyDB (Redis-compatible cache).  
  
## Task  
## Task  
Implement production-ready database and cache clients with connection pooling, health checks, and optimized operations.  
  
## Requirements  
  
### 1. YugabyteDB Client (`internal/database/yugabyte.go`)  
### 1. YugabyteDB Client (`internal/database/yugabyte.go`)  
  
Features needed:  
- Connection pool with pgxpool  
- Multi-host support for HA (topology-aware load balancing)  
- Connection retry with exponential backoff  
- Health check endpoint  
- Cluster info retrieval (nodes, zones, regions)  
- Transaction support with retry logic for serialization errors  
- Read replica support (follower reads)  
- Prepared statement caching  
- Metrics collection (connection pool stats)  
  
Configuration options:  
```go  
type YugabyteConfig struct {  
type YugabyteConfig struct {  
    Hosts           []string  
    Port            int  
    Database        string  
    Database        string  
    User            string  
    Password        string  
    Password        string  
    SSLMode         string  
    MaxConns        int32  
    MaxConns        int32  
    MinConns        int32  
    MinConns        int32  
    MaxConnLifetime time.Duration  
    LoadBalance     bool  
    LoadBalance     bool  
    TopologyKeys    string  *// cloud.region.zone format*  
}  
```  
  
Methods needed:  
- NewYugabyteDB(config) (**YugabyteDB, error)*  
- NewYugabyteDB(config) (**YugabyteDB, error)*  
*- Pool() **pgxpool.Pool  
*- Pool() **pgxpool.Pool  
- Health(ctx) error  
- GetClusterInfo(ctx) (*ClusterInfo, error)  
- ExecuteInTransaction(ctx, fn) error  
- Close()  
  
### 2. DragonflyDB Client (`internal/cache/dragonfly.go`)  
  
Features needed:  
- Connection to single instance or cluster  
- Connection pooling  
- Health check  
- Generic Get/Set with JSON marshaling  
- GetOrSet pattern (cache-aside)  
- Session management (SessionStore)  
- Rate limiting (sliding window)  
- Distributed locking  
- Pub/Sub support  
- Pipeline operations  
- Metrics collection  
  
Methods needed:  
- NewDragonflyDB(config) (**DragonflyDB, error)*  
- Health(ctx) error  
- Set(ctx, key, value, expiration) error  
- Get(ctx, key, dest) error  
- Delete(ctx, keys...) error  
- GetOrSet(ctx, key, dest, expiration, fn) error  
*- NewSessionStore(prefix, ttl) **SessionStore  
*- NewSessionStore(prefix, ttl) **SessionStore  
- NewRateLimiter(prefix) **RateLimiter*  
- NewRateLimiter(prefix) **RateLimiter*  
*- NewLock(key, ttl) **Lock  
*- NewLock(key, ttl) **Lock  
- NewPublisher() **Publisher*  
- NewPublisher() **Publisher*  
*- NewSubscriber(ctx, channels...) **Subscriber  
*- NewSubscriber(ctx, channels...) **Subscriber  
- Close() error  
  
### 3. Repository Layer (`internal/repository/`)  
### 3. Repository Layer (`internal/repository/`)  
  
Create repositories for each domain model:  
- ProjectRepository  
- ApplicationRepository  
- DeploymentRepository  
- ClusterRepository  
- TeamRepository  
- UserRepository  
- AuditLogRepository  
  
Each repository should have:  
- Create, Get, Update, Delete methods  
- List with pagination and filtering  
- Batch operations where applicable  
- Soft delete support  
- Audit logging integration  
  
Example interface:  
```go  
type ApplicationRepository interface {  
type ApplicationRepository interface {  
    Create(ctx context.Context, app *models.Application) error  
    Create(ctx context.Context, app *models.Application) error  
    Get(ctx context.Context, id string) (*models.Application, error)  
    Get(ctx context.Context, id string) (*models.Application, error)  
    GetByProjectID(ctx context.Context, projectID string) ([]*models.Application, error)  
    Update(ctx context.Context, app *models.Application) error  
    Delete(ctx context.Context, id string) error  
    Delete(ctx context.Context, id string) error  
    List(ctx context.Context, opts ListOptions) (*ApplicationList, error)  
    List(ctx context.Context, opts ListOptions) (*ApplicationList, error)  
}  
```  
  
### 4. Database Migrations (`migrations/`)  
  
Create SQL migrations for all tables:  
- Use golang-migrate format (XXXXXX_name.up.sql / .down.sql)  
- Include proper indexes  
- Use YugabyteDB-specific features (tablets, colocated tables)  
- Add foreign key constraints  
- Include audit triggers  
  
Tables needed:  
- teams  
- users  
- team_members (junction)  
- projects  
- applications  
- deployments  
- deployment_revisions  
- clusters  
- cluster_nodes  
- secrets (encrypted)  
- environment_variables  
- audit_logs  
  
### 5. Unit Tests  
### 5. Unit Tests  
  
Create comprehensive tests for:  
- YugabyteDB client (use testcontainers or mock)  
- DragonflyDB client (use testcontainers or mock)  
- Repository operations  
- Connection pool behavior  
- Error handling  
- Retry logic  
  
## Deliverables  
1. internal/database/yugabyte.go - Full YugabyteDB client  
2. internal/cache/dragonfly.go - Full DragonflyDB client  
3. internal/repository/**.go - All repository implementations*  
3. internal/repository/**.go - All repository implementations*  
*4. migrations/**.sql - All database migrations  
*4. migrations/**.sql - All database migrations  
5. internal/database/yugabyte_test.go - Tests  
6. internal/cache/dragonfly_test.go - Tests  
  
## Performance Requirements  
## Performance Requirements  
- Connection pool should handle 1000+ concurrent connections  
- Cache operations < 1ms for local operations  
- Database queries should use prepared statements  
- Implement connection warming on startup  
  
**Prompt 3: NATS Event Bus and Hasura Integration**  
  
  
markdown  
# NorthStack Platform - Event Bus and GraphQL Layer  
  
## Context  
Implementing the event-driven architecture using NATS JetStream and Hasura GraphQL engine for the NorthStack platform.  
  
## Task  
## Task  
Build a robust event bus for inter-service communication and integrate Hasura for GraphQL API capabilities.  
  
## Requirements  
## Requirements  
  
### 1. NATS Event Bus (`internal/events/eventbus.go`)  
### 1. NATS Event Bus (`internal/events/eventbus.go`)  
  
Create a comprehensive event bus with:  
  
****Streams Configuration:****  
- DEPLOYMENTS: Build and deployment events (30-day retention)  
- APPLICATIONS: Application lifecycle events (30-day retention)  
- CLUSTERS: Cluster state changes (7-day retention)  
- AUDIT: Audit trail events (365-day retention)  
- ALERTS: System alerts (7-day retention)  
  
****Features:****  
- Automatic stream creation on startup  
- Durable consumers with acknowledgment  
- Event publishing with subject routing  
- Event subscription with handlers  
- Dead letter queue handling  
- Event replay capability  
- Metrics collection  
  
****Subject Naming Convention:****  
****Subject Naming Convention:****  
```  
{domain}.{action}.{entity_id}  
Example: deployment.started.app-123  
         cluster.node.added.cluster-456  
         audit.user.login.user-789  
```  
  
****Event Types:****  
****Event Types:****  
```go  
type Event struct {  
    ID        string                 `json:"id"`  
    ID        string                 `json:"id"`  
    Type      string                 `json:"type"`  
    Type      string                 `json:"type"`  
    Source    string                 `json:"source"`  
    Subject   string                 `json:"subject"`  
    Subject   string                 `json:"subject"`  
    Timestamp time.Time              `json:"timestamp"`  
    Timestamp time.Time              `json:"timestamp"`  
    Data      map[string]interface{} `json:"data"`  
    Data      map[string]interface{} `json:"data"`  
    Metadata  EventMetadata          `json:"metadata"`  
    Metadata  EventMetadata          `json:"metadata"`  
}  
  
type EventMetadata struct {  
    CorrelationID string `json:"correlationId"`  
    CorrelationID string `json:"correlationId"`  
    CausationID   string `json:"causationId,omitempty"`  
    CausationID   string `json:"causationId,omitempty"`  
    UserID        string `json:"userId,omitempty"`  
    UserID        string `json:"userId,omitempty"`  
    TraceID       string `json:"traceId,omitempty"`  
    TraceID       string `json:"traceId,omitempty"`  
}  
```  
  
****Methods:****  
- NewEventBus(config) (**EventBus, error)*  
- NewEventBus(config) (**EventBus, error)*  
- Publish(ctx, subject, event) error  
- PublishAsync(subject, event) error  
*- Subscribe(subject, handler) (**Subscription, error)  
*- Subscribe(subject, handler) (**Subscription, error)  
- QueueSubscribe(subject, queue, handler) (*Subscription, error)  
- GetHistory(ctx, subject, since, until, limit) ([]Event, error)  
- Health(ctx) error  
- Close() error  
  
### 2. Event Handlers (`internal/events/handlers/`)  
### 2. Event Handlers (`internal/events/handlers/`)  
  
Create handlers for:  
- Deployment events → Update DragonflyDB cache, notify WebSocket clients  
- Build events → Track build progress, trigger downstream deployments  
- Cluster events → Update cluster status, alert on issues  
- Audit events → Persist to YugabyteDB  
  
### 3. Hasura Client (`pkg/hasura/client.go`)  
  
Implement Hasura management client:  
  
****Features:****  
- GraphQL query execution  
- Metadata API operations  
- Schema tracking (tables, relationships)  
- Permission management  
- Action creation  
- Event trigger management  
- Remote schema integration  
- Metadata export/import  
  
****Methods:****  
****Methods:****  
```go  
type HasuraClient interface {  
    *// GraphQL*  
    *// GraphQL*  
    Execute(ctx, query, variables, headers) (*GraphQLResponse, error)  
    Execute(ctx, query, variables, headers) (*GraphQLResponse, error)  
    Subscribe(ctx, query, variables, handler) (*Subscription, error)  
      
    *// Metadata*  
    *// Metadata*  
    TrackTable(ctx, schema, table) error  
    CreateRelationship(ctx, rel *Relationship) error  
    CreateRelationship(ctx, rel *Relationship) error  
    CreatePermission(ctx, perm *Permission) error  
    CreateAction(ctx, action *Action) error  
    CreateAction(ctx, action *Action) error  
    CreateEventTrigger(ctx, trigger *EventTrigger) error  
    CreateEventTrigger(ctx, trigger *EventTrigger) error  
      
    *// Management*  
    ExportMetadata(ctx) (json.RawMessage, error)  
    ExportMetadata(ctx) (json.RawMessage, error)  
    ApplyMetadata(ctx, metadata) error  
    ReloadMetadata(ctx) error  
      
    *// Health*  
    *// Health*  
    Health(ctx) error  
}  
```  
  
### 4. Hasura Metadata Setup  
  
Create initialization code that:  
1. Tracks all platform tables  
2. Creates relationships between tables  
3. Sets up role-based permissions:  
   - admin: Full access  
   - user: Read own resources, write own projects  
   - anonymous: Read public endpoints  
4. Creates event triggers for:  
   - deployments → notify on status change  
   - applications → sync to cache  
   - audit_logs → real-time streaming  
  
### 5. GraphQL Schema Extensions  
### 5. GraphQL Schema Extensions  
  
Create Hasura Actions for operations that need Go logic:  
- deployApplication(id, options) → Deployment  
- rollbackDeployment(id, revision) → Deployment  
- scaleApplication(id, replicas) → Application  
- provisionCluster(config) → Cluster  
  
### 6. WebSocket Event Bridge  
### 6. WebSocket Event Bridge  
  
Bridge NATS events to GraphQL subscriptions:  
```go  
type WebSocketBridge struct {  
type WebSocketBridge struct {  
    eventBus *EventBus  
    hasura   *HasuraClient  
}  
  
*// Bridge NATS events to Hasura live queries*  
*// Bridge NATS events to Hasura live queries*  
func (b *WebSocketBridge) Start(ctx context.Context) error  
func (b *WebSocketBridge) Start(ctx context.Context) error  
```  
  
### 7. Tests  
### 7. Tests  
  
- Event publishing/subscribing tests  
- Event replay tests  
- Hasura metadata operations tests  
- GraphQL query execution tests  
- WebSocket bridge tests  
  
## Deliverables  
## Deliverables  
1. internal/events/eventbus.go - NATS JetStream event bus  
2. internal/events/handlers/*.go - Event handlers  
3. internal/events/types.go - Event type definitions  
4. pkg/hasura/client.go - Hasura client  
5. pkg/hasura/metadata.go - Metadata operations  
6. internal/hasura/setup.go - Initial metadata setup  
7. internal/events/bridge.go - WebSocket bridge  
8. Complete test files  
  
## Integration Requirements  
## Integration Requirements  
- Events should flow: Source → NATS → Handlers → Cache/DB/WebSocket  
- Hasura should receive events via webhooks from event triggers  
- All operations should be traced with correlation IDs  
  
**Prompt 4: Integration Clients (Coolify, Rancher, ArgoCD)**  
  
  
markdown  
# NorthStack Platform - Integration Clients  
  
## Context  
## Context  
Building API clients to integrate with Coolify (CI/CD), Rancher (cluster management), and ArgoCD (GitOps) for the NorthStack unified platform.  
  
## Task  
## Task  
Implement production-ready API clients that abstract the complexity of each tool and provide a consistent interface.  
  
## Requirements  
  
### 1. Coolify Client (`pkg/coolify/client.go`)  
### 1. Coolify Client (`pkg/coolify/client.go`)  
  
****Endpoints to implement:****  
****Endpoints to implement:****  
- Projects: CRUD operations  
- Applications: CRUD, deploy, restart, stop  
- Builds: Trigger, status, logs  
- Preview environments: Create, delete, list  
- Environment variables: CRUD  
- Databases: PostgreSQL, MySQL, Redis provisioning  
- Servers: List, health check  
  
****Features:****  
- HTTP client with retry logic  
- Webhook signature validation  
- Build log streaming (SSE)  
- Automatic token refresh if supported  
- Rate limiting respect  
- Error mapping to domain errors  
  
****Key Methods:****  
****Key Methods:****  
```go  
type CoolifyClient interface {  
type CoolifyClient interface {  
    *// Applications*  
    *// Applications*  
    ListApplications(ctx) ([]*Application, error)  
    GetApplication(ctx, id) (*Application, error)  
    CreateApplication(ctx, req *CreateAppRequest) (*Application, error)  
    UpdateApplication(ctx, id, req *UpdateAppRequest) (*Application, error)  
    UpdateApplication(ctx, id, req *UpdateAppRequest) (*Application, error)  
    DeleteApplication(ctx, id) error  
    DeleteApplication(ctx, id) error  
      
    *// Builds*  
    TriggerBuild(ctx, appID, commitSHA) (*Build, error)  
    GetBuild(ctx, appID, buildID) (*Build, error)  
    GetBuildLogs(ctx, appID, buildID) (io.ReadCloser, error)  
    GetBuildLogs(ctx, appID, buildID) (io.ReadCloser, error)  
    CancelBuild(ctx, appID, buildID) error  
    CancelBuild(ctx, appID, buildID) error  
      
    *// Preview Environments*  
    *// Preview Environments*  
    CreatePreviewEnv(ctx, appID, branch, prNumber) (*PreviewEnv, error)  
    ListPreviewEnvs(ctx, appID) ([]*PreviewEnv, error)  
    ListPreviewEnvs(ctx, appID) ([]*PreviewEnv, error)  
    DeletePreviewEnv(ctx, id) error  
    DeletePreviewEnv(ctx, id) error  
      
    *// Environment Variables*  
    ListEnvVars(ctx, appID) ([]*EnvVar, error)  
    ListEnvVars(ctx, appID) ([]*EnvVar, error)  
    SetEnvVars(ctx, appID, vars []*EnvVar) error  
      
    *// Webhooks*  
    ValidateWebhook(payload []byte, signature string) bool  
    ValidateWebhook(payload []byte, signature string) bool  
    ParseWebhookEvent(payload []byte) (*WebhookEvent, error)  
    ParseWebhookEvent(payload []byte) (*WebhookEvent, error)  
      
    *// Health*  
    *// Health*  
    Health(ctx) error  
    Health(ctx) error  
}  
```  
  
### 2. Rancher Client (`pkg/rancher/client.go`)  
  
****Endpoints to implement:****  
****Endpoints to implement:****  
- Clusters: CRUD, import, provisioning (RKE2)  
- Nodes: List, drain, cordon, delete  
- Namespaces: CRUD  
- Workloads: List, get  
- Projects (Rancher projects): CRUD  
- RBAC: Roles, bindings  
- Fleet: GitRepo, Bundle management  
  
****Features:****  
- API v3 support  
- WebSocket for real-time updates  
- Kubeconfig generation  
- Multi-cluster operations  
- Cluster provisioning with RKE2  
  
****Key Methods:****  
****Key Methods:****  
```go  
type RancherClient interface {  
type RancherClient interface {  
    *// Clusters*  
    ListClusters(ctx) ([]*Cluster, error)  
    ListClusters(ctx) ([]*Cluster, error)  
    GetCluster(ctx, id) (*Cluster, error)  
    CreateCluster(ctx, req *CreateClusterRequest) (*Cluster, error)  
    CreateCluster(ctx, req *CreateClusterRequest) (*Cluster, error)  
    ImportCluster(ctx, name string) (*ImportInfo, error)  
    DeleteCluster(ctx, id) error  
    DeleteCluster(ctx, id) error  
    GetKubeconfig(ctx, clusterID) ([]byte, error)  
    GetKubeconfig(ctx, clusterID) ([]byte, error)  
      
    *// Nodes*  
    *// Nodes*  
    ListNodes(ctx, clusterID) ([]*Node, error)  
    ListNodes(ctx, clusterID) ([]*Node, error)  
    DrainNode(ctx, clusterID, nodeID) error  
    DrainNode(ctx, clusterID, nodeID) error  
    CordonNode(ctx, clusterID, nodeID) error  
    DeleteNode(ctx, clusterID, nodeID) error  
    DeleteNode(ctx, clusterID, nodeID) error  
      
    *// Namespaces*  
    *// Namespaces*  
    ListNamespaces(ctx, clusterID) ([]*Namespace, error)  
    ListNamespaces(ctx, clusterID) ([]*Namespace, error)  
    CreateNamespace(ctx, clusterID, name string) (*Namespace, error)  
      
    *// Fleet GitOps*  
    *// Fleet GitOps*  
    CreateGitRepo(ctx, req *GitRepoRequest) (*GitRepo, error)  
    ListGitRepos(ctx) ([]*GitRepo, error)  
    ListGitRepos(ctx) ([]*GitRepo, error)  
    GetBundleStatus(ctx, gitRepoID) (*BundleStatus, error)  
    GetBundleStatus(ctx, gitRepoID) (*BundleStatus, error)  
      
    *// Health*  
    Health(ctx) error  
    Health(ctx) error  
}  
```  
  
### 3. ArgoCD Client (`pkg/argocd/client.go`)  
  
****Features:****  
- gRPC client using official ArgoCD SDK  
- Application management  
- Sync operations with strategies  
- Rollback support  
- Repository management  
- Project management  
- Notification configuration  
  
****Key Methods:****  
```go  
type ArgoCDClient interface {  
    *// Applications*  
    *// Applications*  
    ListApplications(ctx, project string) ([]*Application, error)  
    ListApplications(ctx, project string) ([]*Application, error)  
    GetApplication(ctx, name string) (*Application, error)  
    CreateApplication(ctx, req *CreateAppRequest) (*Application, error)  
    CreateApplication(ctx, req *CreateAppRequest) (*Application, error)  
    UpdateApplication(ctx, name string, req *UpdateAppRequest) (*Application, error)  
    DeleteApplication(ctx, name string, cascade bool) error  
    DeleteApplication(ctx, name string, cascade bool) error  
      
    *// Sync Operations*  
    Sync(ctx, name string, opts *SyncOptions) (*SyncResult, error)  
    Sync(ctx, name string, opts *SyncOptions) (*SyncResult, error)  
    TerminateSync(ctx, name string) error  
    TerminateSync(ctx, name string) error  
    GetSyncStatus(ctx, name string) (*SyncStatus, error)  
    GetSyncStatus(ctx, name string) (*SyncStatus, error)  
      
    *// Rollback*  
    Rollback(ctx, name string, revisionID int64) (*RollbackResult, error)  
    Rollback(ctx, name string, revisionID int64) (*RollbackResult, error)  
    GetRevisionHistory(ctx, name string) ([]*RevisionInfo, error)  
    GetRevisionHistory(ctx, name string) ([]*RevisionInfo, error)  
      
    *// Repositories*  
    *// Repositories*  
    AddRepository(ctx, repo *Repository) error  
    AddRepository(ctx, repo *Repository) error  
    ListRepositories(ctx) ([]*Repository, error)  
    ListRepositories(ctx) ([]*Repository, error)  
    DeleteRepository(ctx, url string) error  
    DeleteRepository(ctx, url string) error  
      
    *// Projects*  
    *// Projects*  
    CreateProject(ctx, project *Project) error  
    CreateProject(ctx, project *Project) error  
    GetProject(ctx, name string) (*Project, error)  
      
    *// Resource Tree*  
    GetResourceTree(ctx, appName string) (*ResourceTree, error)  
    GetResourceTree(ctx, appName string) (*ResourceTree, error)  
    GetResourceLogs(ctx, appName, podName, container string) (io.ReadCloser, error)  
      
    *// Health*  
    *// Health*  
    Health(ctx) error  
    Health(ctx) error  
}  
```  
  
### 4. RKE2 Manager (`pkg/rke2/manager.go`)  
  
****Features:****  
****Features:****  
- Generate RKE2 server/agent configurations  
- SSH-based node provisioning  
- Cluster bootstrap  
- Node join operations  
- Cluster upgrades  
- Backup/restore operations  
  
****Key Methods:****  
****Key Methods:****  
```go  
type RKE2Manager interface {  
    *// Configuration*  
    GenerateServerConfig(cfg *ClusterConfig) (string, error)  
    GenerateServerConfig(cfg *ClusterConfig) (string, error)  
    GenerateAgentConfig(cfg *NodeConfig) (string, error)  
    GenerateInstallScript(params *InstallParams) string  
      
    *// Provisioning*  
    ProvisionServer(ctx, executor *SSHExecutor, cfg *ClusterConfig) error  
    ProvisionServer(ctx, executor *SSHExecutor, cfg *ClusterConfig) error  
    ProvisionAgent(ctx, executor *SSHExecutor, serverAddr string, token string) error  
    ProvisionAgent(ctx, executor *SSHExecutor, serverAddr string, token string) error  
      
    *// Management*  
    *// Management*  
    GetClusterStatus(ctx, kubeconfig string) (*ClusterStatus, error)  
    GetClusterStatus(ctx, kubeconfig string) (*ClusterStatus, error)  
    UpgradeCluster(ctx, kubeconfig string, version string) error  
      
    *// Backup*  
    CreateEtcdBackup(ctx, kubeconfig string) (*Backup, error)  
    CreateEtcdBackup(ctx, kubeconfig string) (*Backup, error)  
    RestoreEtcdBackup(ctx, kubeconfig string, backupPath string) error  
    RestoreEtcdBackup(ctx, kubeconfig string, backupPath string) error  
}  
```  
  
### 5. Unified Service Aggregator (`internal/services/aggregator.go`)  
### 5. Unified Service Aggregator (`internal/services/aggregator.go`)  
  
Create a service that:  
- Combines data from all sources  
- Caches frequently accessed data in DragonflyDB  
- Publishes events to NATS  
- Handles cross-service transactions  
- Provides unified error handling  
  
### 6. Webhook Handlers  
### 6. Webhook Handlers  
  
Create webhook handlers for:  
- Coolify build events  
- ArgoCD sync events  
- GitHub/GitLab push events  
- Rancher cluster events  
  
Convert webhooks to NATS events.  
  
## Deliverables  
## Deliverables  
1. pkg/coolify/client.go - Full Coolify client  
2. pkg/rancher/client.go - Full Rancher client  
3. pkg/argocd/client.go - Full ArgoCD client  
4. pkg/rke2/manager.go - RKE2 provisioning manager  
5. internal/services/aggregator.go - Unified aggregator  
6. internal/api/webhooks/ - Webhook handlers  
7. Complete test files with mocks  
  
## Error Handling  
- Define domain-specific errors  
- Map API errors to domain errors  
- Include retry logic for transient failures  
- Log all API interactions for debugging  
  
**Prompt 5: REST API and Authentication**  
  
  
markdown  
# NorthStack Platform - REST API Implementation  
  
## Context  
## Context  
Building the REST API layer for NorthStack platform using Go Fiber framework with JWT authentication, rate limiting, and comprehensive middleware.  
  
## Task  
Implement a production-ready REST API that exposes all platform functionality through a clean, documented interface.  
  
## Requirements  
## Requirements  
  
### 1. API Server (`internal/api/server.go`)  
### 1. API Server (`internal/api/server.go`)  
  
****Configuration:****  
****Configuration:****  
- Configurable port and host  
- Graceful shutdown  
- Request timeout  
- Body size limits  
- CORS configuration  
- Compression  
  
****Middleware Stack:****  
****Middleware Stack:****  
1. Recovery (panic handling)  
2. Request ID generation  
3. Structured logging  
4. CORS  
5. Rate limiting  
6. Authentication (where required)  
7. Authorization (RBAC)  
8. Request validation  
9. Response compression  
  
### 2. Authentication (`internal/middleware/auth.go`)  
  
****JWT Implementation:****  
****JWT Implementation:****  
- Access tokens (15 min expiry)  
- Refresh tokens (7 day expiry, stored in DragonflyDB)  
- Token rotation on refresh  
- Revocation support  
- Claims: user_id, email, roles, teams  
  
****Auth Endpoints:****  
****Auth Endpoints:****  
```  
POST /api/v1/auth/register  
POST /api/v1/auth/login  
POST /api/v1/auth/refresh  
POST /api/v1/auth/logout  
POST /api/v1/auth/forgot-password  
POST /api/v1/auth/reset-password  
GET  /api/v1/auth/me  
```  
  
****OAuth2/OIDC Support:****  
****OAuth2/OIDC Support:****  
- GitHub  
- GitLab  
- Google  
- Generic OIDC provider  
  
### 3. API Routes  
  
****Projects:****  
```  
GET    /api/v1/projects  
POST   /api/v1/projects  
GET    /api/v1/projects/:id  
PUT    /api/v1/projects/:id  
DELETE /api/v1/projects/:id  
GET    /api/v1/projects/:id/applications  
GET    /api/v1/projects/:id/members  
POST   /api/v1/projects/:id/members  
DELETE /api/v1/projects/:id/members/:userId  
```  
  
****Applications:****  
```  
GET    /api/v1/applications  
POST   /api/v1/applications  
GET    /api/v1/applications/:id  
PUT    /api/v1/applications/:id  
DELETE /api/v1/applications/:id  
POST   /api/v1/applications/:id/deploy  
POST   /api/v1/applications/:id/rollback  
POST   /api/v1/applications/:id/restart  
POST   /api/v1/applications/:id/stop  
POST   /api/v1/applications/:id/scale  
GET    /api/v1/applications/:id/deployments  
GET    /api/v1/applications/:id/logs  
GET    /api/v1/applications/:id/metrics  
GET    /api/v1/applications/:id/events  
```  
  
****Deployments:****  
****Deployments:****  
```  
GET    /api/v1/deployments  
GET    /api/v1/deployments/:id  
POST   /api/v1/deployments/:id/promote  
POST   /api/v1/deployments/:id/abort  
GET    /api/v1/deployments/:id/logs  
```  
  
****Clusters:****  
```  
GET    /api/v1/clusters  
POST   /api/v1/clusters  
GET    /api/v1/clusters/:id  
PUT    /api/v1/clusters/:id  
DELETE /api/v1/clusters/:id  
GET    /api/v1/clusters/:id/nodes  
POST   /api/v1/clusters/:id/nodes  
DELETE /api/v1/clusters/:id/nodes/:nodeId  
GET    /api/v1/clusters/:id/kubeconfig  
GET    /api/v1/clusters/:id/metrics  
POST   /api/v1/clusters/:id/upgrade  
```  
  
****Secrets:****  
****Secrets:****  
```  
GET    /api/v1/secrets  
POST   /api/v1/secrets  
GET    /api/v1/secrets/:id  
PUT    /api/v1/secrets/:id  
DELETE /api/v1/secrets/:id  
```  
  
****Teams:****  
****Teams:****  
```  
GET    /api/v1/teams  
POST   /api/v1/teams  
GET    /api/v1/teams/:id  
PUT    /api/v1/teams/:id  
DELETE /api/v1/teams/:id  
GET    /api/v1/teams/:id/members  
POST   /api/v1/teams/:id/members  
DELETE /api/v1/teams/:id/members/:userId  
```  
  
****Events:****  
```  
GET    /api/v1/events  
GET    /api/v1/events/stream (SSE)  
```  
  
****Webhooks (Public):****  
```  
POST   /webhooks/github  
POST   /webhooks/gitlab  
POST   /webhooks/coolify  
POST   /webhooks/argocd  
```  
  
### 4. Request/Response Models  
  
Create DTOs for all endpoints:  
- Request validation using go-playground/validator  
- Response formatting with consistent structure  
- Pagination support  
- Filtering and sorting  
- Error responses with codes  
  
****Standard Response Format:****  
****Standard Response Format:****  
```go  
type Response struct {  
    Success bool        `json:"success"`  
    Success bool        `json:"success"`  
    Data    interface{} `json:"data,omitempty"`  
    Error   *ErrorInfo  `json:"error,omitempty"`  
    Meta    *MetaInfo   `json:"meta,omitempty"`  
    Meta    *MetaInfo   `json:"meta,omitempty"`  
}  
  
type ErrorInfo struct {  
    Code    string            `json:"code"`  
    Message string            `json:"message"`  
    Details map[string]string `json:"details,omitempty"`  
}  
  
type MetaInfo struct {  
type MetaInfo struct {  
    Page       int   `json:"page,omitempty"`  
    Page       int   `json:"page,omitempty"`  
    PerPage    int   `json:"perPage,omitempty"`  
    PerPage    int   `json:"perPage,omitempty"`  
    TotalCount int64 `json:"totalCount,omitempty"`  
    TotalPages int   `json:"totalPages,omitempty"`  
    TotalPages int   `json:"totalPages,omitempty"`  
}  
```  
  
### 5. Rate Limiting (`internal/middleware/ratelimit.go`)  
### 5. Rate Limiting (`internal/middleware/ratelimit.go`)  
  
****Features:****  
- Per-user rate limiting  
- Per-endpoint rate limiting  
- Sliding window algorithm  
- DragonflyDB backend  
- Response headers (X-RateLimit-*)  
  
****Configuration:****  
****Configuration:****  
```go  
type RateLimitConfig struct {  
type RateLimitConfig struct {  
    Enabled       bool  
    Enabled       bool  
    RequestsPerMin int  
    RequestsPerMin int  
    BurstSize     int  
    ByUser        bool  
    ByUser        bool  
    ByIP          bool  
    ByIP          bool  
    ByEndpoint    bool  
    Whitelist     []string  
    Whitelist     []string  
}  
```  
  
### 6. WebSocket Handler (`internal/api/websocket.go`)  
### 6. WebSocket Handler (`internal/api/websocket.go`)  
  
****Features:****  
- Real-time event streaming  
- Per-resource subscriptions  
- Authentication via query param or first message  
- Heartbeat/ping-pong  
- Reconnection handling  
  
****Protocol:****  
****Protocol:****  
```json  
*// Subscribe*  
*// Subscribe*  
{"type": "subscribe", "channel": "application.app-123"}  
  
*// Unsubscribe*  
{"type": "unsubscribe", "channel": "application.app-123"}  
  
*// Event*  
{"type": "event", "channel": "application.app-123", "data": {...}}  
{"type": "event", "channel": "application.app-123", "data": {...}}  
```  
  
### 7. OpenAPI Documentation  
### 7. OpenAPI Documentation  
  
Generate OpenAPI 3.0 spec:  
- Use swaggo/swag for annotation-based generation  
- Include all endpoints  
- Request/response schemas  
- Authentication requirements  
- Error responses  
  
### 8. Health Endpoints  
### 8. Health Endpoints  
```  
GET /health          - Basic health check  
GET /health/ready    - Readiness (all deps healthy)  
GET /health/live     - Liveness  
GET /metrics         - Prometheus metrics  
```  
  
## Deliverables  
1. internal/api/server.go - Main API server  
2. internal/api/routes.go - Route registration  
3. internal/api/handlers/**.go - All endpoint handlers*  
*4. internal/middleware/**.go - All middleware  
*4. internal/middleware/**.go - All middleware  
5. internal/api/dto/*.go - Request/response DTOs  
6. internal/api/websocket.go - WebSocket handler  
7. docs/swagger.yaml - OpenAPI spec  
8. Complete test files  
  
## Security Requirements  
## Security Requirements  
- Input validation on all endpoints  
- SQL injection prevention (parameterized queries)  
- XSS prevention (output encoding)  
- CSRF protection for state-changing operations  
- Secure headers (HSTS, CSP, etc.)  
- Audit logging for sensitive operations  
  
**Prompt 6: Frontend Portal (React/TypeScript)**  
  
  
markdown  
# NorthStack Platform - Frontend Portal  
  
## Context  
Building the unified dashboard for NorthStack platform using React, TypeScript, and TanStack Query for data fetching.  
  
## Task  
Create a modern, responsive dashboard that provides a single-pane-of-glass view for managing applications, clusters, and deployments.  
  
## Requirements  
  
### 1. Project Setup  
### 1. Project Setup  
  
****Technology Stack:****  
****Technology Stack:****  
- React 18+ with TypeScript  
- Vite for build tooling  
- TanStack Query for data fetching  
- TanStack Router for routing  
- Zustand for global state  
- Tailwind CSS for styling  
- shadcn/ui for components  
- Recharts for visualizations  
- React Hook Form + Zod for forms  
  
****Directory Structure:****  
```  
frontend/  
├── src/  
│   ├── api/              # API client functions  
│   ├── components/       # Reusable components  
│   │   ├── ui/          # shadcn/ui components  
│   │   ├── layout/      # Layout components  
│   │   └── features/    # Feature-specific components  
│   ├── hooks/           # Custom hooks  
│   ├── pages/           # Page components  
│   ├── stores/          # Zustand stores  
│   ├── types/           # TypeScript types  
│   ├── utils/           # Utility functions  
│   ├── App.tsx  
│   └── main.tsx  
├── public/  
├── index.html  
├── tailwind.config.js  
├── tsconfig.json  
└── vite.config.ts  
```  
  
### 2. Core Features  
### 2. Core Features  
  
****Dashboard Page:****  
****Dashboard Page:****  
- Overview statistics (apps, clusters, deployments)  
- Recent activity feed  
- Quick actions  
- System health indicators  
- Resource usage charts  
  
****Applications Page:****  
****Applications Page:****  
- List view with filtering and sorting  
- Grid/table toggle  
- Search functionality  
- Status indicators  
- Quick deploy button  
- Bulk actions  
  
****Application Detail Page:****  
****Application Detail Page:****  
- Overview tab (status, endpoints, metrics)  
- Deployments tab (history, rollback)  
- Logs tab (real-time streaming)  
- Settings tab (env vars, scaling)  
- Metrics tab (CPU, memory, requests)  
  
****Clusters Page:****  
- Cluster list with health status  
- Node information  
- Resource utilization  
- Cluster creation wizard  
  
****Deployments Page:****  
****Deployments Page:****  
- Deployment timeline  
- Canary/Blue-green progress  
- Rollback controls  
- Deployment comparison  
  
### 3. API Client (`src/api/`)  
  
Create typed API client:  
```typescript  
*// src/api/client.ts*  
const api = {  
const api = {  
  applications: {  
  applications: {  
    list: (params?: ListParams) => Promise,  
    list: (params?: ListParams) => Promise,  
    get: (id: string) => Promise,  
    create: (data: CreateApplicationInput) => Promise,  
    create: (data: CreateApplicationInput) => Promise,  
    update: (id: string, data: UpdateApplicationInput) => Promise,  
    delete: (id: string) => Promise,  
    delete: (id: string) => Promise,  
    deploy: (id: string, options?: DeployOptions) => Promise,  
    deploy: (id: string, options?: DeployOptions) => Promise,  
    rollback: (id: string, revision: string) => Promise,  
    rollback: (id: string, revision: string) => Promise,  
  },  
  clusters: { ... },  
  clusters: { ... },  
  deployments: { ... },  
  *// etc.*  
  *// etc.*  
}  
```  
  
### 4. Real-time Updates  
  
****WebSocket Integration:****  
```typescript  
*// src/hooks/useWebSocket.ts*  
function useWebSocket(channels: string[], onMessage: (event: Event) => void)  
function useWebSocket(channels: string[], onMessage: (event: Event) => void)  
  
*// src/hooks/useApplicationEvents.ts*  
*// src/hooks/useApplicationEvents.ts*  
function useApplicationEvents(appId: string) {  
function useApplicationEvents(appId: string) {  
  *// Subscribe to application-specific events*  
  *// Subscribe to application-specific events*  
  *// Update TanStack Query cache on events*  
  *// Update TanStack Query cache on events*  
}  
```  
  
****Features:****  
****Features:****  
- Automatic reconnection  
- Event-based cache invalidation  
- Optimistic updates  
- Toast notifications for important events  
  
### 5. Components to Build  
### 5. Components to Build  
  
****Layout:****  
- Sidebar navigation  
- Header with user menu  
- Breadcrumbs  
- Mobile-responsive drawer  
  
****Data Display:****  
- DataTable with sorting, filtering, pagination  
- StatusBadge (Running, Failed, Pending, etc.)  
- MetricsChart (line, area, bar)  
- LogViewer (virtualized, searchable)  
- Terminal (for SSH/exec)  
- Timeline (deployment history)  
- ResourceUsage (CPU, memory gauges)  
  
****Forms:****  
- ApplicationForm (create/edit)  
- ClusterForm (create/edit)  
- SecretForm (create/edit)  
- DeploymentOptionsForm  
  
****Feedback:****  
- Toast notifications  
- Confirmation dialogs  
- Loading skeletons  
- Error boundaries  
- Empty states  
  
### 6. State Management  
### 6. State Management  
  
****Zustand Stores:****  
```typescript  
*// src/stores/authStore.ts*  
interface AuthStore {  
  user: User | null;  
  token: string | null;  
  token: string | null;  
  login: (credentials: Credentials) => Promise;  
  login: (credentials: Credentials) => Promise;  
  logout: () => void;  
  logout: () => void;  
  refreshToken: () => Promise;  
}  
  
*// src/stores/uiStore.ts*  
interface UIStore {  
  sidebarOpen: boolean;  
  sidebarOpen: boolean;  
  theme: 'light' | 'dark' | 'system';  
  toggleSidebar: () => void;  
  setTheme: (theme: Theme) => void;  
  setTheme: (theme: Theme) => void;  
}  
```  
  
### 7. GraphQL Integration (Hasura)  
### 7. GraphQL Integration (Hasura)  
  
Create Apollo Client setup for GraphQL:  
- Queries for complex data fetching  
- Subscriptions for real-time updates  
- Mutations for data changes  
```typescript  
*// src/api/graphql/queries.ts*  
*// src/api/graphql/queries.ts*  
export const GET_APPLICATION_WITH_DEPLOYMENTS = gql`  
export const GET_APPLICATION_WITH_DEPLOYMENTS = gql`  
  query GetApplicationWithDeployments($id: uuid!) {  
    applications_by_pk(id: $id) {  
      id  
      name  
      status  
      deployments(order_by: { created_at: desc }, limit: 10) {  
        id  
        revision  
        status  
        created_at  
      }  
    }  
  }  
`;  
```  
  
### 8. Theming  
  
Implement dark/light mode:  
- System preference detection  
- Manual toggle  
- Persistent preference  
- Consistent color palette  
  
### 9. Testing  
  
- Unit tests with Vitest  
- Component tests with Testing Library  
- E2E tests with Playwright  
- MSW for API mocking  
  
## Deliverables  
## Deliverables  
1. Complete React application structure  
2. All page components  
3. Reusable UI component library  
4. API client with types  
5. WebSocket integration  
6. GraphQL setup  
7. State management  
8. Authentication flow  
9. Test setup with examples  
  
## Design Requirements  
## Design Requirements  
- Mobile-first responsive design  
- Accessibility (WCAG 2.1 AA)  
- Loading states for all async operations  
- Error handling with user-friendly messages  
- Keyboard navigation support  
  
**Prompt 7: Deployment and Infrastructure**  
  
  
markdown  
# NorthStack Platform - Deployment Infrastructure  
  
## Context  
## Context  
Creating the deployment infrastructure for NorthStack platform including Kubernetes manifests, Helm charts, and installation automation.  
  
## Task  
Build production-ready deployment configurations for self-hosting the NorthStack platform.  
  
## Requirements  
  
### 1. Kubernetes Manifests (`deployments/kubernetes/`)  
### 1. Kubernetes Manifests (`deployments/kubernetes/`)  
  
Create base manifests for:  
- NorthStack API (Deployment, Service, HPA)  
- Frontend Portal (Deployment, Service)  
- YugabyteDB (StatefulSet via Helm subchart reference)  
- DragonflyDB (StatefulSet)  
- NATS (StatefulSet)  
- Hasura (Deployment, Service)  
- Traefik Ingress Controller  
- Prometheus + Grafana (reference kube-prometheus-stack)  
  
Use Kustomize for:  
- Base configuration  
- Overlays: development, staging, production  
- Environment-specific patches  
  
### 2. Helm Chart (`deployments/helm/northstack/`)  
  
****Chart Structure:****  
****Chart Structure:****  
```  
northstack/  
├── Chart.yaml  
├── values.yaml  
├── values-production.yaml  
├── templates/  
│   ├── _helpers.tpl  
│   ├── deployment-api.yaml  
│   ├── deployment-portal.yaml  
│   ├── deployment-hasura.yaml  
│   ├── statefulset-dragonfly.yaml  
│   ├── statefulset-nats.yaml  
│   ├── service-*.yaml  
│   ├── ingress.yaml  
│   ├── configmap.yaml  
│   ├── secret.yaml  
│   ├── hpa.yaml  
│   ├── pdb.yaml  
│   ├── networkpolicy.yaml  
│   ├── serviceaccount.yaml  
│   ├── rbac.yaml  
│   └── servicemonitor.yaml  
├── charts/               # Subcharts  
│   └── yugabyte/        # YugabyteDB subchart  
└── README.md  
```  
  
****values.yaml Structure:****  
****values.yaml Structure:****  
```yaml  
global:  
  imageRegistry: ""  
  imagePullSecrets: []  
  storageClass: ""  
  
api:  
  replicaCount: 3  
  replicaCount: 3  
  image:  
    repository: northstack/api  
    tag: latest  
  resources:  
    requests:  
      cpu: 100m  
      memory: 256Mi  
    limits:  
      cpu: 500m  
      memory: 512Mi  
  autoscaling:  
    enabled: true  
    minReplicas: 2  
    maxReplicas: 10  
    targetCPUUtilization: 70  
    targetCPUUtilization: 70  
  
portal:  
  replicaCount: 2  
  image:  
    repository: northstack/portal  
    tag: latest  
  
yugabyte:  
  enabled: true  
  enabled: true  
  replicas: 3  
  resource:  
    master:  
      requests:  
        cpu: 500m  
        memory: 1Gi  
    tserver:  
      requests:  
        cpu: 1  
        memory: 2Gi  
  storage:  
    master:  
      size: 10Gi  
    tserver:  
      size: 100Gi  
  
dragonfly:  
  enabled: true  
  replicas: 3  
  resources:  
    requests:  
      cpu: 500m  
      memory: 1Gi  
  storage:  
    size: 10Gi  
  
nats:  
  enabled: true  
  enabled: true  
  replicas: 3  
  replicas: 3  
  jetstream:  
    enabled: true  
    enabled: true  
    storage: 10Gi  
  
hasura:  
  enabled: true  
  enabled: true  
  adminSecret: ""  *# Required*  
  adminSecret: ""  *# Required*  
  jwtSecret: ""    *# Required*  
  
ingress:  
  enabled: true  
  enabled: true  
  className: traefik  
  annotations: {}  
  hosts:  
    - host: northstack.example.com  
      paths:  
        - path: /api  
          service: api  
        - path: /graphql  
          service: hasura  
        - path: /  
          service: portal  
  tls:  
    enabled: true  
    enabled: true  
    secretName: northstack-tls  
  
monitoring:  
  enabled: true  
  enabled: true  
  serviceMonitor:  
    enabled: true  
    enabled: true  
  grafanaDashboards:  
    enabled: true  
    enabled: true  
```  
  
### 3. Docker Images  
### 3. Docker Images  
  
****Dockerfiles:****  
****Dockerfiles:****  
```dockerfile  
*# Dockerfile.api*  
*# Dockerfile.api*  
FROM golang:1.22-alpine AS builder  
WORKDIR /app  
COPY go.mod go.sum ./  
RUN go mod download  
COPY . .  
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api  
  
FROM gcr.io/distroless/static-debian12  
COPY --from=builder /api /api  
COPY --from=builder /app/migrations /migrations  
USER nonroot:nonroot  
EXPOSE 8080 9090  
ENTRYPOINT ["/api"]  
```  
```dockerfile  
*# Dockerfile.portal*  
*# Dockerfile.portal*  
FROM node:20-alpine AS builder  
WORKDIR /app  
COPY package*.json ./  
RUN npm ci  
COPY . .  
RUN npm run build  
  
FROM nginx:alpine  
COPY --from=builder /app/dist /usr/share/nginx/html  
COPY nginx.conf /etc/nginx/nginx.conf  
EXPOSE 80  
CMD ["nginx", "-g", "daemon off;"]  
```  
  
### 4. Installation Script (`scripts/install.sh`)  
### 4. Installation Script (`scripts/install.sh`)  
  
Create comprehensive installation script:  
- Prerequisites check (kubectl, helm, etc.)  
- RKE2 cluster bootstrap (optional)  
- Namespace creation  
- Secret generation (passwords, JWT keys)  
- Helm chart installation  
- Post-install configuration  
- Health verification  
- Output connection details  
  
### 5. ArgoCD ApplicationSet  
### 5. ArgoCD ApplicationSet  
  
For GitOps-based deployment:  
```yaml  
apiVersion: argoproj.io/v1alpha1  
kind: ApplicationSet  
metadata:  
  name: northstack  
  namespace: argocd  
spec:  
  generators:  
    - list:  
        elements:  
          - cluster: production  
            url: https://kubernetes.default.svc  
            values:  
              valuesFile: values-production.yaml  
          - cluster: staging  
            url: https://staging.example.com  
            values:  
              valuesFile: values-staging.yaml  
  template:  
    metadata:  
      name: 'northstack-{{cluster}}'  
      name: 'northstack-{{cluster}}'  
    spec:  
      project: default  
      source:  
        repoURL: https://github.com/org/northstack-gitops  
        targetRevision: HEAD  
        path: deployments/helm/northstack  
        helm:  
          valueFiles:  
            - '{{values.valuesFile}}'  
            - '{{values.valuesFile}}'  
      destination:  
        server: '{{url}}'  
        namespace: northstack  
      syncPolicy:  
        automated:  
          prune: true  
          selfHeal: true  
          selfHeal: true  
```  
  
### 6. Monitoring Configuration  
  
****Prometheus ServiceMonitors:****  
- NorthStack API metrics  
- YugabyteDB metrics  
- DragonflyDB metrics  
- NATS metrics  
- Hasura metrics  
  
****Grafana Dashboards:****  
- Platform overview  
- Application performance  
- Cluster health  
- Database performance  
- Cache performance  
  
### 7. Backup Configuration  
### 7. Backup Configuration  
  
Create CronJobs for:  
- YugabyteDB backup to S3  
- NATS JetStream backup  
- Configuration backup  
  
### 8. Network Policies  
### 8. Network Policies  
  
Implement least-privilege network policies:  
- API → YugabyteDB, DragonflyDB, NATS, Hasura  
- Hasura → YugabyteDB, DragonflyDB  
- Portal → API, Hasura  
- Ingress → API, Portal, Hasura  
  
## Deliverables  
## Deliverables  
1. deployments/kubernetes/ - Complete Kustomize structure  
2. deployments/helm/northstack/ - Complete Helm chart  
3. Dockerfile.api and Dockerfile.portal  
4. scripts/install.sh - Installation script  
5. ArgoCD ApplicationSet  
6. Monitoring configurations  
7. Backup CronJobs  
8. Network Policies  
9. Comprehensive README with installation guide  
  
## Production Requirements  
- High availability (3+ replicas)  
- Resource limits on all containers  
- Pod disruption budgets  
- Horizontal pod autoscaling  
- Node affinity/anti-affinity  
- Security contexts (non-root, read-only filesystem)  
- Network policies  
- Secret management (external-secrets or sealed-secrets)  
  
**Summary: Complete Claude Code Prompt Set**  

| Prompt # | Focus Area | Key Deliverables |
| -------- | ---------------------- | ------------------------------------------------- |
| 1 | Project Initialization | Structure, config, models, Makefile |
| 2 | Data Layer | YugabyteDB, DragonflyDB, repositories, migrations |
| 3 | Events & GraphQL | NATS event bus, Hasura client, WebSocket bridge |
| 4 | Integration Clients | Coolify, Rancher, ArgoCD, RKE2 clients |
| 5 | REST API | Fiber API, JWT auth, rate limiting, WebSocket |
| 6 | Frontend | React portal, real-time updates, components |
| 7 | Deployment | Helm chart, Kubernetes manifests, installation |
  
Each prompt is designed to be self-contained but builds on the previous work. Use them sequentially with Claude Code to build the complete platform.  
Would you like me to:  
1. Expand any specific prompt with more detail?  
2. Create additional prompts for specific features (CLI tool, Vault integration, etc.)?  
3. Generate the actual database migration SQL files?  
4. Create the Grafana dashboard JSON configurations?  
  
  
  
  
  
  
  
