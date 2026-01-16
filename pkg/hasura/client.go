package hasura

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides methods for interacting with Hasura GraphQL Engine
type Client struct {
	endpoint    string
	adminSecret string
	httpClient  *http.Client
}

// Config holds Hasura client configuration
type Config struct {
	Endpoint    string
	AdminSecret string
	Timeout     time.Duration
}

// NewClient creates a new Hasura client
func NewClient(cfg *Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		endpoint:    cfg.Endpoint,
		adminSecret: cfg.AdminSecret,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
}

// Query executes a GraphQL query
func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	req := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	return c.execute(ctx, "/v1/graphql", req, result)
}

// Mutation executes a GraphQL mutation
func (c *Client) Mutation(ctx context.Context, mutation string, variables map[string]interface{}, result interface{}) error {
	req := GraphQLRequest{
		Query:     mutation,
		Variables: variables,
	}
	return c.execute(ctx, "/v1/graphql", req, result)
}

// MetadataRequest represents a Hasura metadata API request
type MetadataRequest struct {
	Type    string      `json:"type"`
	Version int         `json:"version,omitempty"`
	Args    interface{} `json:"args"`
}

// TrackTable tracks a database table in Hasura
func (c *Client) TrackTable(ctx context.Context, schema, table string) error {
	req := MetadataRequest{
		Type:    "pg_track_table",
		Version: 1,
		Args: map[string]interface{}{
			"source": "default",
			"table": map[string]string{
				"schema": schema,
				"name":   table,
			},
		},
	}
	return c.executeMetadata(ctx, req, nil)
}

// UntrackTable untracks a table from Hasura
func (c *Client) UntrackTable(ctx context.Context, schema, table string) error {
	req := MetadataRequest{
		Type:    "pg_untrack_table",
		Version: 1,
		Args: map[string]interface{}{
			"source": "default",
			"table": map[string]string{
				"schema": schema,
				"name":   table,
			},
			"cascade": true,
		},
	}
	return c.executeMetadata(ctx, req, nil)
}

// CreateRelationship creates a relationship between tables
type RelationshipArgs struct {
	Name   string                 `json:"name"`
	Source string                 `json:"source"`
	Table  map[string]string      `json:"table"`
	Using  map[string]interface{} `json:"using"`
}

func (c *Client) CreateObjectRelationship(ctx context.Context, args RelationshipArgs) error {
	req := MetadataRequest{
		Type:    "pg_create_object_relationship",
		Version: 1,
		Args:    args,
	}
	return c.executeMetadata(ctx, req, nil)
}

func (c *Client) CreateArrayRelationship(ctx context.Context, args RelationshipArgs) error {
	req := MetadataRequest{
		Type:    "pg_create_array_relationship",
		Version: 1,
		Args:    args,
	}
	return c.executeMetadata(ctx, req, nil)
}

// Permission types
type PermissionArgs struct {
	Source     string                 `json:"source"`
	Table      map[string]string      `json:"table"`
	Role       string                 `json:"role"`
	Permission map[string]interface{} `json:"permission"`
}

// CreateSelectPermission creates a select permission
func (c *Client) CreateSelectPermission(ctx context.Context, args PermissionArgs) error {
	req := MetadataRequest{
		Type:    "pg_create_select_permission",
		Version: 1,
		Args:    args,
	}
	return c.executeMetadata(ctx, req, nil)
}

// CreateInsertPermission creates an insert permission
func (c *Client) CreateInsertPermission(ctx context.Context, args PermissionArgs) error {
	req := MetadataRequest{
		Type:    "pg_create_insert_permission",
		Version: 1,
		Args:    args,
	}
	return c.executeMetadata(ctx, req, nil)
}

// CreateUpdatePermission creates an update permission
func (c *Client) CreateUpdatePermission(ctx context.Context, args PermissionArgs) error {
	req := MetadataRequest{
		Type:    "pg_create_update_permission",
		Version: 1,
		Args:    args,
	}
	return c.executeMetadata(ctx, req, nil)
}

// CreateDeletePermission creates a delete permission
func (c *Client) CreateDeletePermission(ctx context.Context, args PermissionArgs) error {
	req := MetadataRequest{
		Type:    "pg_create_delete_permission",
		Version: 1,
		Args:    args,
	}
	return c.executeMetadata(ctx, req, nil)
}

// EventTriggerArgs defines an event trigger
type EventTriggerArgs struct {
	Name      string                 `json:"name"`
	Source    string                 `json:"source"`
	Table     map[string]string      `json:"table"`
	Webhook   string                 `json:"webhook"`
	Insert    *OperationSpec         `json:"insert,omitempty"`
	Update    *OperationSpec         `json:"update,omitempty"`
	Delete    *OperationSpec         `json:"delete,omitempty"`
	Headers   []map[string]string    `json:"headers,omitempty"`
	RetryConf map[string]interface{} `json:"retry_conf,omitempty"`
}

type OperationSpec struct {
	Columns []string `json:"columns"`
}

// CreateEventTrigger creates an event trigger
func (c *Client) CreateEventTrigger(ctx context.Context, args EventTriggerArgs) error {
	req := MetadataRequest{
		Type:    "pg_create_event_trigger",
		Version: 1,
		Args:    args,
	}
	return c.executeMetadata(ctx, req, nil)
}

// DeleteEventTrigger deletes an event trigger
func (c *Client) DeleteEventTrigger(ctx context.Context, name, source string) error {
	req := MetadataRequest{
		Type:    "pg_delete_event_trigger",
		Version: 1,
		Args: map[string]string{
			"name":   name,
			"source": source,
		},
	}
	return c.executeMetadata(ctx, req, nil)
}

// ActionDefinition defines a Hasura action
type ActionDefinition struct {
	Name       string                 `json:"name"`
	Definition map[string]interface{} `json:"definition"`
}

// CreateAction creates a Hasura action
func (c *Client) CreateAction(ctx context.Context, action ActionDefinition) error {
	req := MetadataRequest{
		Type:    "create_action",
		Version: 1,
		Args:    action,
	}
	return c.executeMetadata(ctx, req, nil)
}

// ReloadMetadata reloads Hasura metadata
func (c *Client) ReloadMetadata(ctx context.Context) error {
	req := MetadataRequest{
		Type:    "reload_metadata",
		Version: 1,
		Args:    map[string]bool{"reload_remote_schemas": true},
	}
	return c.executeMetadata(ctx, req, nil)
}

// ExportMetadata exports all Hasura metadata
func (c *Client) ExportMetadata(ctx context.Context) (map[string]interface{}, error) {
	req := MetadataRequest{
		Type: "export_metadata",
		Args: map[string]interface{}{},
	}
	var result map[string]interface{}
	if err := c.executeMetadata(ctx, req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ApplyMetadata applies Hasura metadata
func (c *Client) ApplyMetadata(ctx context.Context, metadata map[string]interface{}) error {
	req := MetadataRequest{
		Type:    "replace_metadata",
		Version: 2,
		Args: map[string]interface{}{
			"allow_inconsistent_metadata": false,
			"metadata":                    metadata,
		},
	}
	return c.executeMetadata(ctx, req, nil)
}

// HealthCheck checks if Hasura is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint+"/healthz", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hasura health check failed: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) execute(ctx context.Context, path string, request GraphQLRequest, result interface{}) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.adminSecret != "" {
		req.Header.Set("x-hasura-admin-secret", c.adminSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("graphql error: %s", gqlResp.Errors[0].Message)
	}

	if result != nil && gqlResp.Data != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

func (c *Client) executeMetadata(ctx context.Context, request MetadataRequest, result interface{}) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/v1/metadata", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.adminSecret != "" {
		req.Header.Set("x-hasura-admin-secret", c.adminSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("metadata API error: %s", string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
