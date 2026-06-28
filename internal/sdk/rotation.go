package mazevault

import (
	"fmt"
	"net/http"
	"time"
)

// RotationResourceDetail is the embedded rotation resource read model returned
// alongside rotation config API responses.
type RotationResourceDetail struct {
	ID                   string     `json:"id"`
	Enabled              bool       `json:"enabled"`
	ManualOnly           bool       `json:"manual_only"`
	ScheduleExpression   string     `json:"schedule_expression"`
	RotationIntervalDays *int       `json:"rotation_interval_days"`
	StatusSummary        string     `json:"status_summary"`
	NextDueAt            *time.Time `json:"next_due_at"`
	EnvironmentScope     string     `json:"environment_scope"`
}

// RotationConfigDetail represents a full rotation config as returned by the API.
// NOTE: The backend SecretRotationConfigResponse uses camelCase for several fields;
// JSON tags here match the actual wire format exactly.
type RotationConfigDetail struct {
	ID       string `json:"id"`
	SecretID string `json:"secret_id"`
	Enabled  bool   `json:"enabled"`
	Schedule string `json:"schedule"`
	// RotationIntervalDays is serialised as camelCase by the backend.
	RotationIntervalDays int `json:"rotationIntervalDays"`
	// NotificationEmails is serialised as camelCase by the backend.
	NotificationEmails []string `json:"notificationEmails"`
	// PostActions is serialised as camelCase by the backend.
	PostActions       []PostRotationActionWF `json:"postActions,omitempty"`
	TargetEnvironment string                 `json:"target_environment"`
	BackupStrategy    string                 `json:"backup_strategy"`
	TransactionMode   string                 `json:"transaction_mode"`
	Status            string                 `json:"status"`
	LastError         string                 `json:"last_error,omitempty"`
	LastRotatedAt     *time.Time             `json:"lastRotatedAt,omitempty"`
	NextRotationAt    *time.Time             `json:"nextRotationAt,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	// Resource holds the canonical RotationResource read model embedded in the response.
	Resource *RotationResourceDetail `json:"resource,omitempty"`
}

// CreateRotationConfigRequest represents the request body for creating a rotation config.
type CreateRotationConfigRequest struct {
	SecretID             string                 `json:"secret_id"`
	Enabled              bool                   `json:"enabled"`
	Schedule             string                 `json:"schedule,omitempty"`
	RotationIntervalDays int                    `json:"rotation_interval_days,omitempty"`
	NotificationEmails   []string               `json:"notification_emails,omitempty"`
	PostActions          []PostRotationActionWF `json:"post_actions,omitempty"`
}

// UpdateRotationConfigRequest represents the request body for updating a rotation config.
type UpdateRotationConfigRequest struct {
	Enabled              bool                   `json:"enabled"`
	Schedule             string                 `json:"schedule,omitempty"`
	RotationIntervalDays int                    `json:"rotation_interval_days,omitempty"`
	NotificationEmails   []string               `json:"notification_emails,omitempty"`
	PostActions          []PostRotationActionWF `json:"post_actions,omitempty"`
}

// CreateRotationConfig creates a new rotation config for a secret.
func (c *Client) CreateRotationConfig(req *CreateRotationConfigRequest) (*RotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodPost, "/api/v1/rotation/configs", req)
	if err != nil {
		return nil, err
	}
	var result RotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRotationConfig retrieves a rotation config by ID.
func (c *Client) GetRotationConfig(id string) (*RotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/rotation/configs/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var result RotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRotationConfig updates an existing rotation config.
func (c *Client) UpdateRotationConfig(id string, req *UpdateRotationConfigRequest) (*RotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/rotation/configs/%s", id), req)
	if err != nil {
		return nil, err
	}
	var result RotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRotationConfig deletes a rotation config by ID.
func (c *Client) DeleteRotationConfig(id string) error {
	httpReq, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/rotation/configs/%s", id), nil)
	if err != nil {
		return err
	}
	return c.Do(httpReq, nil)
}

// RotationWorkflowDetail represents a rotation workflow (backed by same RotationConfig on the server side,
// but exposed with richer workflow step and integration data for the Terraform provider).
type RotationWorkflowDetail struct {
	ID                   string                 `json:"id"`
	SecretID             string                 `json:"secret_id"`
	Environment          string                 `json:"environment"`
	Enabled              bool                   `json:"enabled"`
	Schedule             string                 `json:"schedule"`
	RotationIntervalDays int                    `json:"rotation_interval_days"`
	RotationStrategy     string                 `json:"rotation_strategy,omitempty"`
	TargetEnvironment    string                 `json:"target_environment,omitempty"`
	GracePeriodMinutes   int                    `json:"grace_period_minutes,omitempty"`
	ResourceKind         string                 `json:"resource_kind,omitempty"`
	LinkedIntegrations   []string               `json:"linked_integrations,omitempty"`
	PostRotationActions  []PostRotationActionWF `json:"post_rotation_actions,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// PostRotationActionWF represents a single post-rotation action in a workflow.
type PostRotationActionWF struct {
	Type      string            `json:"type"`
	Config    map[string]string `json:"config,omitempty"`
	Order     int               `json:"order,omitempty"`
	OnFailure string            `json:"on_failure,omitempty"`
	// GatewayID pins this action to a specific gateway instead of using environment-based routing.
	GatewayID string `json:"gateway_id,omitempty"`
	// TargetEnvironment overrides the environment context used when resolving the target gateway.
	TargetEnvironment string `json:"target_environment,omitempty"`
}

// CreateRotationWorkflowRequest creates a rotation workflow via the rotation/configs endpoint.
type CreateRotationWorkflowRequest struct {
	SecretID             string                 `json:"secret_id"`
	Enabled              bool                   `json:"enabled"`
	Schedule             string                 `json:"schedule,omitempty"`
	RotationIntervalDays int                    `json:"rotation_interval_days,omitempty"`
	RotationStrategy     string                 `json:"rotation_strategy,omitempty"`
	TargetEnvironment    string                 `json:"target_environment,omitempty"`
	GracePeriodMinutes   int                    `json:"grace_period_minutes,omitempty"`
	ResourceKind         string                 `json:"resource_kind,omitempty"`
	LinkedIntegrations   []string               `json:"linked_integrations,omitempty"`
	PostRotationActions  []PostRotationActionWF `json:"post_rotation_actions,omitempty"`
}

// CreateRotationWorkflow creates a new rotation workflow.
func (c *Client) CreateRotationWorkflow(req *CreateRotationWorkflowRequest) (*RotationWorkflowDetail, error) {
	httpReq, err := c.newRequest(http.MethodPost, "/api/v1/rotation/configs", req)
	if err != nil {
		return nil, err
	}
	var result RotationWorkflowDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRotationWorkflow retrieves a rotation workflow by ID.
func (c *Client) GetRotationWorkflow(id string) (*RotationWorkflowDetail, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/rotation/configs/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var result RotationWorkflowDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRotationWorkflow updates a rotation workflow.
func (c *Client) UpdateRotationWorkflow(id string, req *CreateRotationWorkflowRequest) (*RotationWorkflowDetail, error) {
	httpReq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/rotation/configs/%s", id), req)
	if err != nil {
		return nil, err
	}
	var result RotationWorkflowDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRotationWorkflow deletes a rotation workflow by ID.
func (c *Client) DeleteRotationWorkflow(id string) error {
	httpReq, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/rotation/configs/%s", id), nil)
	if err != nil {
		return err
	}
	return c.Do(httpReq, nil)
}

// ─────────────────────────────────────────────────────────────────────────────
// Certificate Rotation Config
// ─────────────────────────────────────────────────────────────────────────────

// CertificateRotationConfigDetail is the full rotation config for a certificate
// as returned by GET /api/v1/certificates/:id/rotation/config.
type CertificateRotationConfigDetail struct {
	ID                  string                  `json:"id"`
	CertificateID       string                  `json:"certificate_id"`
	ConfigInitialized   bool                    `json:"config_initialized"`
	Enabled             bool                    `json:"enabled"`
	Schedule            string                  `json:"schedule"`
	RenewalLeadDays     int                     `json:"renewal_lead_days"`
	MaxRetryAttempts    int                     `json:"max_retry_attempts"`
	RetryDelaySeconds   int                     `json:"retry_delay_seconds"`
	TimeoutMinutes      int                     `json:"timeout_minutes"`
	NotificationEmails  []string                `json:"notification_emails,omitempty"`
	PostRotationActions []PostRotationActionWF  `json:"post_rotation_actions,omitempty"`
	LastRotatedAt       *time.Time              `json:"last_rotated_at,omitempty"`
	NextRotationAt      *time.Time              `json:"next_rotation_at,omitempty"`
	Resource            *RotationResourceDetail `json:"resource,omitempty"`
}

// UpdateCertRotationConfigRequest is the request body for
// PUT /api/v1/certificates/:id/rotation/config.
type UpdateCertRotationConfigRequest struct {
	Enabled             *bool                  `json:"enabled"`
	Schedule            string                 `json:"schedule,omitempty"`
	RenewalLeadDays     *int                   `json:"renewal_lead_days,omitempty"`
	MaxRetryAttempts    *int                   `json:"max_retry_attempts,omitempty"`
	RetryDelaySeconds   *int                   `json:"retry_delay_seconds,omitempty"`
	TimeoutMinutes      *int                   `json:"timeout_minutes,omitempty"`
	NotificationEmails  []string               `json:"notification_emails,omitempty"`
	PostRotationActions []PostRotationActionWF `json:"post_rotation_actions,omitempty"`
}

// GetCertificateRotationConfig retrieves the rotation config for a certificate.
func (c *Client) GetCertificateRotationConfig(certID string) (*CertificateRotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/certificates/%s/rotation/config", certID), nil)
	if err != nil {
		return nil, err
	}
	var result CertificateRotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateCertificateRotationConfig upserts the rotation config for a certificate.
func (c *Client) UpdateCertificateRotationConfig(certID string, req *UpdateCertRotationConfigRequest) (*CertificateRotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/certificates/%s/rotation/config", certID), req)
	if err != nil {
		return nil, err
	}
	var result CertificateRotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Entra Credential Rotation Config
// ─────────────────────────────────────────────────────────────────────────────

// EntraRotationConfigDetail is the full rotation config for an Entra credential
// as returned by GET /api/v1/entra/credentials/:id/rotation-config.
type EntraRotationConfigDetail struct {
	CredentialID             string                  `json:"credential_id"`
	RotationEnabled          bool                    `json:"rotation_enabled"`
	RotationDaysBeforeExpiry int                     `json:"rotation_days_before_expiry"`
	GracePeriodDays          int                     `json:"grace_period_days"`
	KVIntegrationIDs         []string                `json:"kv_integration_ids,omitempty"`
	SecretName               string                  `json:"secret_name,omitempty"`
	SpringEndpoints          []string                `json:"spring_endpoints,omitempty"`
	WebhookURLs              []string                `json:"webhook_urls,omitempty"`
	PostRotationActions      []PostRotationActionWF  `json:"post_rotation_actions,omitempty"`
	StagedRotationEnabled    bool                    `json:"staged_rotation_enabled"`
	SoakWindowHours          int                     `json:"soak_window_hours,omitempty"`
	LastRotatedAt            *time.Time              `json:"last_rotated_at,omitempty"`
	Resource                 *RotationResourceDetail `json:"resource,omitempty"`
}

// ConfigureEntraRotationRequest is the shared request body for both
// POST /api/v1/entra/credentials/:id/rotation-config  (create) and
// PUT  /api/v1/entra/credentials/:id/rotation-config  (update).
type ConfigureEntraRotationRequest struct {
	RotationEnabled          bool                   `json:"rotation_enabled"`
	RotationDaysBeforeExpiry int                    `json:"rotation_days_before_expiry,omitempty"`
	GracePeriodDays          int                    `json:"grace_period_days,omitempty"`
	KVIntegrationIDs         []string               `json:"kv_integration_ids,omitempty"`
	SecretName               string                 `json:"secret_name,omitempty"`
	SpringEndpoints          []string               `json:"spring_endpoints,omitempty"`
	WebhookURLs              []string               `json:"webhook_urls,omitempty"`
	PostRotationActions      []PostRotationActionWF `json:"post_rotation_actions,omitempty"`
	StagedRotationEnabled    *bool                  `json:"staged_rotation_enabled,omitempty"`
	SoakWindowHours          *int                   `json:"soak_window_hours,omitempty"`
}

// GetEntraRotationConfig retrieves the rotation config for an Entra credential.
func (c *Client) GetEntraRotationConfig(credentialID string) (*EntraRotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/entra/credentials/%s/rotation-config", credentialID), nil)
	if err != nil {
		return nil, err
	}
	var result EntraRotationConfigDetail
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfigureEntraRotation creates a rotation config for an Entra credential (POST).
func (c *Client) ConfigureEntraRotation(credentialID string, req *ConfigureEntraRotationRequest) (*EntraRotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/api/v1/entra/credentials/%s/rotation-config", credentialID), req)
	if err != nil {
		return nil, err
	}
	// POST returns {"status": "rotation configured", "config": {...}}
	var wrapper struct {
		Config EntraRotationConfigDetail `json:"config"`
	}
	if err := c.Do(httpReq, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Config, nil
}

// UpdateEntraRotationConfig updates the rotation config for an Entra credential (PUT).
func (c *Client) UpdateEntraRotationConfig(credentialID string, req *ConfigureEntraRotationRequest) (*EntraRotationConfigDetail, error) {
	httpReq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/api/v1/entra/credentials/%s/rotation-config", credentialID), req)
	if err != nil {
		return nil, err
	}
	// PUT returns {"status": "rotation configuration updated"}; re-fetch to get current state.
	if err := c.Do(httpReq, nil); err != nil {
		return nil, err
	}
	return c.GetEntraRotationConfig(credentialID)
}

// ─────────────────────────────────────────────────────────────────────────────
// Rotation Resources (canonical platform read model)
// ─────────────────────────────────────────────────────────────────────────────

// RotationResourceItem is one entry from the rotation resources list endpoint.
type RotationResourceItem struct {
	ID               string     `json:"id"`
	ResourceKind     string     `json:"resource_kind"`
	ResourceID       string     `json:"resource_id"`
	ProjectID        string     `json:"project_id,omitempty"`
	EnvironmentScope string     `json:"environment_scope,omitempty"`
	DisplayName      string     `json:"display_name"`
	Enabled          bool       `json:"enabled"`
	ManualOnly       bool       `json:"manual_only"`
	StatusSummary    string     `json:"status_summary"`
	NextDueAt        *time.Time `json:"next_due_at,omitempty"`
	LastExecutionID  string     `json:"last_execution_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ListRotationResourcesResponse wraps the rotation resources list response.
type ListRotationResourcesResponse struct {
	Resources []RotationResourceItem `json:"resources"`
}

// ListRotationResources lists rotation resources, optionally filtered by kind and
// environment scope.  Pass empty strings to omit the respective filter.
func (c *Client) ListRotationResources(kind, environmentScope string) ([]RotationResourceItem, error) {
	path := "/api/v1/rotation/resources"
	sep := "?"
	if kind != "" {
		path += sep + "kind=" + kind
		sep = "&"
	}
	if environmentScope != "" {
		path += sep + "environment_scope=" + environmentScope
	}

	httpReq, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result ListRotationResourcesResponse
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return result.Resources, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Rotation Resource History
// ─────────────────────────────────────────────────────────────────────────────

// RotationResourceHistoryItem is a single execution record in the rotation history.
type RotationResourceHistoryItem struct {
	ID          string     `json:"id"`
	ConfigID    string     `json:"config_id,omitempty"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// ListRotationResourceHistoryResponse wraps the history list response.
type ListRotationResourceHistoryResponse struct {
	Executions []RotationResourceHistoryItem `json:"executions"`
}

// GetRotationResourceHistory fetches the execution history for a rotation resource.
func (c *Client) GetRotationResourceHistory(kind, resourceID string) ([]RotationResourceHistoryItem, error) {
	path := fmt.Sprintf("/api/v1/rotation/resources/%s/%s/history", kind, resourceID)
	httpReq, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result ListRotationResourceHistoryResponse
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return result.Executions, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Project Rotation Configs
// ─────────────────────────────────────────────────────────────────────────────

// ProjectRotationConfigItem is one entry from the project rotation configs list.
type ProjectRotationConfigItem struct {
	ID                   string     `json:"id"`
	SecretID             string     `json:"secret_id"`
	Enabled              bool       `json:"enabled"`
	Schedule             string     `json:"schedule,omitempty"`
	RotationIntervalDays int        `json:"rotation_interval_days,omitempty"`
	Status               string     `json:"status,omitempty"`
	LastRotatedAt        *time.Time `json:"last_rotated_at,omitempty"`
	NextRotationAt       *time.Time `json:"next_rotation_at,omitempty"`
}

// ListProjectRotationConfigsResponse wraps the project rotation configs list.
type ListProjectRotationConfigsResponse struct {
	Configs []ProjectRotationConfigItem `json:"configs"`
}

// ListProjectRotationConfigs lists rotation configs for a project.
func (c *Client) ListProjectRotationConfigs(projectID string) ([]ProjectRotationConfigItem, error) {
	httpReq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/rotation-configs", projectID), nil)
	if err != nil {
		return nil, err
	}
	var result ListProjectRotationConfigsResponse
	if err := c.Do(httpReq, &result); err != nil {
		return nil, err
	}
	return result.Configs, nil
}
