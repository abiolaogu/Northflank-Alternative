// Package workflow provides the deployment state machine and workflow engine.
// This is the core orchestration logic that coordinates builds, deployments,
// and manages the lifecycle of services across the platform.
package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openpaas/platform-orchestrator/internal/domain"
	"github.com/openpaas/platform-orchestrator/pkg/logger"
)

// DeploymentState represents the current state of a deployment workflow
type DeploymentState string

const (
	StateIdle             DeploymentState = "idle"
	StateBuildQueued      DeploymentState = "build_queued"
	StateBuilding         DeploymentState = "building"
	StateBuildComplete    DeploymentState = "build_complete"
	StateBuildFailed      DeploymentState = "build_failed"
	StateDeployQueued     DeploymentState = "deploy_queued"
	StateDeploying        DeploymentState = "deploying"
	StateDeployComplete   DeploymentState = "deploy_complete"
	StateDeployFailed     DeploymentState = "deploy_failed"
	StateRollingBack      DeploymentState = "rolling_back"
	StateRollbackComplete DeploymentState = "rollback_complete"
)

// DeploymentEvent represents events that trigger state transitions
type DeploymentEvent string

const (
	EventTriggerBuild     DeploymentEvent = "trigger_build"
	EventBuildStarted     DeploymentEvent = "build_started"
	EventBuildSucceeded   DeploymentEvent = "build_succeeded"
	EventBuildFailed      DeploymentEvent = "build_failed"
	EventTriggerDeploy    DeploymentEvent = "trigger_deploy"
	EventDeployStarted    DeploymentEvent = "deploy_started"
	EventDeploySucceeded  DeploymentEvent = "deploy_succeeded"
	EventDeployFailed     DeploymentEvent = "deploy_failed"
	EventTriggerRollback  DeploymentEvent = "trigger_rollback"
	EventRollbackComplete DeploymentEvent = "rollback_complete"
	EventCancel           DeploymentEvent = "cancel"
)

// DeploymentWorkflow represents a deployment workflow instance
type DeploymentWorkflow struct {
	ID           uuid.UUID
	ServiceID    uuid.UUID
	ProjectID    uuid.UUID
	ClusterID    uuid.UUID
	State        DeploymentState
	BuildID      *uuid.UUID
	DeploymentID *uuid.UUID
	Version      string
	PrevVersion  string
	Error        string
	StartedAt    time.Time
	UpdatedAt    time.Time
	Metadata     map[string]interface{}
}

// StateMachine manages deployment workflow state transitions
type StateMachine struct {
	mu          sync.RWMutex
	workflows   map[uuid.UUID]*DeploymentWorkflow
	ciAdapter   domain.CIAdapter
	gitOps      domain.GitOpsAdapter
	eventBus    domain.EventBus
	serviceRepo domain.ServiceRepository
	logger      *logger.Logger
	transitions map[DeploymentState]map[DeploymentEvent]DeploymentState
}

// NewStateMachine creates a new state machine
func NewStateMachine(
	ciAdapter domain.CIAdapter,
	gitOps domain.GitOpsAdapter,
	eventBus domain.EventBus,
	serviceRepo domain.ServiceRepository,
	log *logger.Logger,
) *StateMachine {
	sm := &StateMachine{
		workflows:   make(map[uuid.UUID]*DeploymentWorkflow),
		ciAdapter:   ciAdapter,
		gitOps:      gitOps,
		eventBus:    eventBus,
		serviceRepo: serviceRepo,
		logger:      log,
	}

	sm.initTransitions()
	return sm
}

// initTransitions initializes the valid state transitions
func (sm *StateMachine) initTransitions() {
	sm.transitions = map[DeploymentState]map[DeploymentEvent]DeploymentState{
		StateIdle: {
			EventTriggerBuild:  StateBuildQueued,
			EventTriggerDeploy: StateDeployQueued,
		},
		StateBuildQueued: {
			EventBuildStarted: StateBuilding,
			EventCancel:       StateIdle,
		},
		StateBuilding: {
			EventBuildSucceeded: StateBuildComplete,
			EventBuildFailed:    StateBuildFailed,
			EventCancel:         StateIdle,
		},
		StateBuildComplete: {
			EventTriggerDeploy: StateDeployQueued,
			EventTriggerBuild:  StateBuildQueued,
		},
		StateBuildFailed: {
			EventTriggerBuild: StateBuildQueued,
		},
		StateDeployQueued: {
			EventDeployStarted: StateDeploying,
			EventCancel:        StateBuildComplete,
		},
		StateDeploying: {
			EventDeploySucceeded: StateDeployComplete,
			EventDeployFailed:    StateDeployFailed,
			EventTriggerRollback: StateRollingBack,
		},
		StateDeployComplete: {
			EventTriggerBuild:    StateBuildQueued,
			EventTriggerDeploy:   StateDeployQueued,
			EventTriggerRollback: StateRollingBack,
		},
		StateDeployFailed: {
			EventTriggerRollback: StateRollingBack,
			EventTriggerBuild:    StateBuildQueued,
			EventTriggerDeploy:   StateDeployQueued,
		},
		StateRollingBack: {
			EventRollbackComplete: StateRollbackComplete,
			EventDeployFailed:     StateDeployFailed,
		},
		StateRollbackComplete: {
			EventTriggerBuild:  StateBuildQueued,
			EventTriggerDeploy: StateDeployQueued,
		},
	}
}

// CreateWorkflow creates a new deployment workflow
func (sm *StateMachine) CreateWorkflow(ctx context.Context, serviceID, projectID, clusterID uuid.UUID) (*DeploymentWorkflow, error) {
	workflow := &DeploymentWorkflow{
		ID:        uuid.New(),
		ServiceID: serviceID,
		ProjectID: projectID,
		ClusterID: clusterID,
		State:     StateIdle,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	sm.mu.Lock()
	sm.workflows[workflow.ID] = workflow
	sm.mu.Unlock()

	sm.logger.Info().
		Str("workflow_id", workflow.ID.String()).
		Str("service_id", serviceID.String()).
		Msg("Deployment workflow created")

	return workflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (sm *StateMachine) GetWorkflow(id uuid.UUID) (*DeploymentWorkflow, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	wf, exists := sm.workflows[id]
	return wf, exists
}

// GetWorkflowByService retrieves the active workflow for a service
func (sm *StateMachine) GetWorkflowByService(serviceID uuid.UUID) (*DeploymentWorkflow, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, wf := range sm.workflows {
		if wf.ServiceID == serviceID && wf.State != StateIdle {
			return wf, true
		}
	}
	return nil, false
}

// ProcessEvent processes an event and transitions the workflow state
func (sm *StateMachine) ProcessEvent(ctx context.Context, workflowID uuid.UUID, event DeploymentEvent, data map[string]interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	workflow, exists := sm.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Check if transition is valid
	transitions, ok := sm.transitions[workflow.State]
	if !ok {
		return fmt.Errorf("no transitions from state: %s", workflow.State)
	}

	newState, ok := transitions[event]
	if !ok {
		return fmt.Errorf("invalid transition: %s -> %s", workflow.State, event)
	}

	oldState := workflow.State
	workflow.State = newState
	workflow.UpdatedAt = time.Now()

	// Update workflow data based on event
	if data != nil {
		if buildID, ok := data["build_id"].(uuid.UUID); ok {
			workflow.BuildID = &buildID
		}
		if deployID, ok := data["deployment_id"].(uuid.UUID); ok {
			workflow.DeploymentID = &deployID
		}
		if version, ok := data["version"].(string); ok {
			workflow.Version = version
		}
		if errMsg, ok := data["error"].(string); ok {
			workflow.Error = errMsg
		}
	}

	sm.logger.Info().
		Str("workflow_id", workflowID.String()).
		Str("event", string(event)).
		Str("old_state", string(oldState)).
		Str("new_state", string(newState)).
		Msg("Workflow state transition")

	// Execute side effects based on new state
	go sm.executeSideEffects(ctx, workflow, oldState, newState)

	return nil
}

// executeSideEffects performs actions based on state transitions
func (sm *StateMachine) executeSideEffects(ctx context.Context, workflow *DeploymentWorkflow, oldState, newState DeploymentState) {
	switch newState {
	case StateBuilding:
		sm.updateServiceStatus(ctx, workflow.ServiceID, domain.ServiceStatusBuilding)

	case StateBuildComplete:
		sm.publishEvent(ctx, "build.completed", workflow)

	case StateBuildFailed:
		sm.updateServiceStatus(ctx, workflow.ServiceID, domain.ServiceStatusFailed)
		sm.publishEvent(ctx, "build.failed", workflow)

	case StateDeploying:
		sm.updateServiceStatus(ctx, workflow.ServiceID, domain.ServiceStatusDeploying)
		sm.publishEvent(ctx, "deploy.started", workflow)

	case StateDeployComplete:
		sm.updateServiceStatus(ctx, workflow.ServiceID, domain.ServiceStatusRunning)
		sm.publishEvent(ctx, "deploy.completed", workflow)

	case StateDeployFailed:
		sm.updateServiceStatus(ctx, workflow.ServiceID, domain.ServiceStatusFailed)
		sm.publishEvent(ctx, "deploy.failed", workflow)

	case StateRollbackComplete:
		sm.updateServiceStatus(ctx, workflow.ServiceID, domain.ServiceStatusRunning)
		sm.publishEvent(ctx, "rollback.completed", workflow)
	}
}

// updateServiceStatus updates the service status in the repository
func (sm *StateMachine) updateServiceStatus(ctx context.Context, serviceID uuid.UUID, status domain.ServiceStatus) {
	if err := sm.serviceRepo.UpdateStatus(ctx, serviceID, status); err != nil {
		sm.logger.Error().Err(err).Str("service_id", serviceID.String()).Msg("Failed to update service status")
	}
}

// publishEvent publishes a workflow event to the event bus
func (sm *StateMachine) publishEvent(ctx context.Context, eventType string, workflow *DeploymentWorkflow) {
	event := &domain.Event{
		Type:   eventType,
		Source: "workflow-engine",
		Data: map[string]interface{}{
			"workflow_id": workflow.ID.String(),
			"service_id":  workflow.ServiceID.String(),
			"project_id":  workflow.ProjectID.String(),
			"state":       string(workflow.State),
			"version":     workflow.Version,
		},
	}

	if workflow.BuildID != nil {
		event.Data["build_id"] = workflow.BuildID.String()
	}
	if workflow.DeploymentID != nil {
		event.Data["deployment_id"] = workflow.DeploymentID.String()
	}
	if workflow.Error != "" {
		event.Data["error"] = workflow.Error
	}

	if err := sm.eventBus.Publish(ctx, eventType, event); err != nil {
		sm.logger.Error().Err(err).Str("event_type", eventType).Msg("Failed to publish event")
	}
}

// TriggerBuildAndDeploy is a convenience method to start a full build+deploy workflow
func (sm *StateMachine) TriggerBuildAndDeploy(ctx context.Context, service *domain.Service, clusterID uuid.UUID) (*DeploymentWorkflow, error) {
	workflow, err := sm.CreateWorkflow(ctx, service.ID, service.ProjectID, clusterID)
	if err != nil {
		return nil, err
	}

	// Trigger build
	if err := sm.ProcessEvent(ctx, workflow.ID, EventTriggerBuild, nil); err != nil {
		return nil, err
	}

	// Start the actual build
	build, err := sm.ciAdapter.TriggerBuild(ctx, service, service.BuildSource)
	if err != nil {
		sm.ProcessEvent(ctx, workflow.ID, EventBuildFailed, map[string]interface{}{"error": err.Error()})
		return workflow, err
	}

	// Update workflow with build info
	sm.ProcessEvent(ctx, workflow.ID, EventBuildStarted, map[string]interface{}{"build_id": build.ID})

	return workflow, nil
}

// CleanupOldWorkflows removes completed workflows older than the retention period
func (sm *StateMachine) CleanupOldWorkflows(retention time.Duration) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-retention)
	removed := 0

	for id, wf := range sm.workflows {
		if wf.UpdatedAt.Before(cutoff) && (wf.State == StateIdle || wf.State == StateDeployComplete || wf.State == StateRollbackComplete) {
			delete(sm.workflows, id)
			removed++
		}
	}

	if removed > 0 {
		sm.logger.Info().Int("count", removed).Msg("Cleaned up old workflows")
	}

	return removed
}
